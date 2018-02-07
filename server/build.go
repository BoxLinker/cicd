package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/drone/drone/model"

	"github.com/cabernety/gopkg/httplib"
	"github.com/gorilla/mux"

	"github.com/BoxLinker/cicd/models"
	"github.com/BoxLinker/cicd/pipeline/rpc"
	"github.com/BoxLinker/cicd/pubsub"
	"github.com/BoxLinker/cicd/queue"
)

// GetBuild 根据 repo 和 build_number 获取 build 信息
func (s *Server) GetBuild(w http.ResponseWriter, r *http.Request) {
	buildNum, _ := strconv.Atoi(mux.Vars(r)["number"])
	repo := r.Context().Value("repo").(*models.Repo)
	build, err := s.Manager.Store().GetBuildNumber(repo, buildNum)
	if err != nil {
		httplib.Resp(w, httplib.STATUS_NOT_FOUND, nil, fmt.Sprintf("build not found: %v", err))
		return
	}
	httplib.Resp(w, httplib.STATUS_OK, build)
}

// QueryBuild 查询 build 列表
func (s *Server) QueryBuild(w http.ResponseWriter, r *http.Request) {
	repo := r.Context().Value("repo").(*models.Repo)
	pc := httplib.ParsePageConfig(r)
	builds := s.Manager.Store().QueryBuild(repo, &pc)
	pc.TotalCount = s.Manager.Store().BuildCount(repo)
	httplib.Resp(w, httplib.STATUS_OK, pc.FormatOutput(builds))
}

// PostBuild 根据传入的 build number 生成一次新的构建
func (s *Server) PostBuild(w http.ResponseWriter, r *http.Request) {
	repo := r.Context().Value("repo").(*models.Repo)
	num, _ := strconv.Atoi(mux.Vars(r)["number"])
	store := s.Manager.Store()
	build, err := store.GetBuildNumber(repo, num)
	if err != nil {
		httplib.Resp(w, httplib.STATUS_NOT_FOUND, nil, fmt.Sprintf("build not found: %v", err))
		return
	}

	switch build.Status {
	case models.StatusRunning,
		models.StatusPending,
		models.StatusDeclined,
		models.StatusBlocked,
		models.StatusError:
		httplib.Resp(w, httplib.STATUS_FAILED, nil, fmt.Sprintf("当前构建处于 %s 状态中，构建失败", build.Status))
		return
	}

	conf, err := store.ConfigLoad(build.ConfigID)
	if err != nil {
		httplib.Resp(w, httplib.STATUS_NOT_FOUND, nil, fmt.Sprintf("没有找到对应的构建配置文件."))
		return
	}

	build.ID = 0
	build.Number = 0
	build.Parent = num
	build.Status = models.StatusPending
	build.Started = 0
	build.Finished = 0
	build.Enqueued = time.Now().UTC().Unix()
	build.Error = ""

	err = store.CreateBuild(build)
	if err != nil {
		httplib.Resp(w, httplib.STATUS_INTERNAL_SERVER_ERR, nil, fmt.Sprintf("构建出错: %v", err))
		return
	}

	// Read query string parameters into buildParams, exclude reserved params
	var buildParams = map[string]string{}
	for key, val := range r.URL.Query() {
		switch key {
		case "fork", "event", "deploy_to":
		default:
			// We only accept string literals, because build parameters will be
			// injected as environment variables
			buildParams[key] = val[0]
		}
	}

	// get the previous build so that we can send
	// on status change notifications
	last, _ := store.GetBuildLastBefore(repo, build.Branch, build.ID)
	secs, err := store.SecretList(repo)
	if err != nil {
		logrus.Errorf("PostBuild secret list err: %v", err)
	}
	regs, err := store.RegistryList(repo)
	if err != nil {
		logrus.Errorf("PostBuild registry list err: %v", err)
	}

	b := builder{
		Repo: repo,
		Curr: build,
		Last: last,
		Secs: secs,
		Regs: regs,
		Link: httplib.GetURL(r),
		Yaml: conf.Data,
		Envs: buildParams,
	}

	items, err := b.Build()
	if err != nil {
		build.Status = models.StatusError
		build.Created = time.Now().Unix()
		build.Finished = build.Started
		build.Error = err.Error()
		httplib.Resp(w, httplib.STATUS_FAILED, build)
		return
	}

	var pcounter = len(items)
	for _, item := range items {
		build.Procs = append(build.Procs, item.Proc)
		item.Proc.BuildID = build.ID

		for _, stage := range item.Config.Stages {
			var gid int
			for _, step := range stage.Steps {
				pcounter++
				if gid == 0 {
					gid = pcounter
				}
				proc := &models.Proc{
					BuildID: build.ID,
					Name:    step.Alias,
					PID:     pcounter,
					PPID:    item.Proc.PID,
					PGID:    gid,
					State:   model.StatusPending,
				}
				build.Procs = append(build.Procs, proc)
			}
		}
	}

	err = store.ProcCreate(build.Procs)
	if err != nil {
		logrus.Errorf("cannot restart %s#%d: %s", repo.FullName, build.Number, err)
		build.Status = models.StatusError
		build.Started = time.Now().Unix()
		build.Finished = build.Started
		build.Error = err.Error()
		httplib.Resp(w, httplib.STATUS_FAILED, build)
		return
	}

	httplib.Resp(w, httplib.STATUS_OK, build)

	// publish topic
	message := pubsub.Message{
		Labels: map[string]string{
			"repo":    repo.FullName,
			"private": strconv.FormatBool(repo.IsPrivate),
		},
	}
	buildCopy := *build
	buildCopy.Procs = models.Tree(buildCopy.Procs)
	message.Data, _ = json.Marshal(models.Event{
		Type:  models.Enqueued,
		Repo:  *repo,
		Build: buildCopy,
	})
	s.Manager.Pubsub().Publish(r.Context(), "topic/events", message)

	for _, item := range items {
		task := new(queue.Task)
		task.ID = fmt.Sprint(item.Proc.ID)
		task.Labels = map[string]string{}
		for k, v := range item.Labels {
			task.Labels[k] = v
		}
		task.Labels["platform"] = item.Platform
		task.Labels["repo"] = b.Repo.FullName

		task.Data, _ = json.Marshal(rpc.Pipeline{
			ID:      fmt.Sprint(item.Proc.ID),
			Config:  item.Config,
			Timeout: b.Repo.Timeout,
		})
		logrus.Debugf("logs open: %s", task.ID)
		s.Manager.Logs().Open(context.Background(), task.ID)
		s.Manager.Queue().Push(context.Background(), task)
	}
}
