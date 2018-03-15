package server

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/BoxLinker/cicd/models"
	"github.com/BoxLinker/cicd/modules/envsubst"
	"github.com/BoxLinker/cicd/modules/linter"
	"github.com/BoxLinker/cicd/modules/token"
	"github.com/BoxLinker/cicd/pipeline/backend"
	"github.com/BoxLinker/cicd/pipeline/frontend"
	"github.com/BoxLinker/cicd/pipeline/frontend/yaml"
	"github.com/BoxLinker/cicd/pipeline/frontend/yaml/compiler"
	"github.com/BoxLinker/cicd/pipeline/frontend/yaml/matrix"
	"github.com/BoxLinker/cicd/pipeline/rpc"
	"github.com/BoxLinker/cicd/pubsub"
	"github.com/BoxLinker/cicd/queue"
	"github.com/Sirupsen/logrus"
	"github.com/cabernety/gopkg/httplib"
	"github.com/gorilla/mux"
)

var skipRe = regexp.MustCompile(`\[(?i:ci *skip|skip *ci)\]`)

func (s *Server) Hook(w http.ResponseWriter, r *http.Request) {
	scmType := mux.Vars(r)["scm"]
	if !models.SCMExists(scmType) {
		http.Error(w, "bad scm type", http.StatusBadRequest)
		return
	}

	remote := s.Manager.GetSCM(scmType)
	tmpRepo, build, err := remote.Hook(r)
	if err != nil {
		logrus.Errorf("failure to parse hook. %s", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	if build == nil {
		w.WriteHeader(200)
		return
	}
	if tmpRepo == nil {
		logrus.Errorf("failure to ascertain repo from hook.")
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	// 如果 commit message 里的信息有类似 skip ci 等字样，那就忽略此次提交
	skipMatch := skipRe.FindString(build.Message)
	if len(skipMatch) > 0 {
		logrus.Infof("ignoring hook. %s found in %s.", skipMatch, build.Message)
		w.WriteHeader(204)
		return
	}

	repo, err := s.Manager.Store().GetRepoOwnerName(s.getUserInfo(r), tmpRepo.Owner, tmpRepo.Name, scmType)
	if err != nil {
		logrus.Errorf("failed to find repo %s/%s from hook. %s", tmpRepo.Owner, tmpRepo.Name, err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if !repo.IsActive {
		msg := fmt.Sprintf("ignoring hook. %s/%s is inactive.", tmpRepo.Owner, tmpRepo.Name)
		logrus.Errorln(msg)
		http.Error(w, msg, http.StatusNoContent)
		return
	}

	parsed, err := token.ParseRequest(r, func(t *token.Token) (string, error) {
		return repo.Hash, nil
	})
	if err != nil {
		logrus.Errorf("failed to parse token from hook for %s. %s", repo.FullName, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if parsed.Text != repo.FullName {
		logrus.Errorf("failed to verify token from hook. Expected %s, got %s", repo.FullName, parsed.Text)
		http.Error(w, "failed to verify token from hook.", http.StatusForbidden)
		return
	}

	if repo.UserID == 0 {
		logrus.Warnf("ignoring hook. repo %s has no owner.", repo.FullName)
		w.WriteHeader(204)
		return
	}

	var skipped = true
	if (build.Event == models.EventPush && repo.AllowPush) ||
		(build.Event == models.EventPull && repo.AllowPull) ||
		(build.Event == models.EventDeploy && repo.AllowDeploy) ||
		(build.Event == models.EventTag && repo.AllowTag) {
		skipped = false
	}

	if skipped {
		logrus.Infof("ignoring hook. repo %s is disabled.", repo.FullName)
		w.WriteHeader(204)
		return
	}

	user, err := s.Manager.GetUserByIDAndSCM(repo.UserID, repo.SCM)
	if err != nil {
		logrus.Errorf("failed to find repo (%s) by owner (%s) and scm (%s)", repo.FullName, repo.UserID, repo.SCM)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// todo refresh token if needed

	// 从 repo 里获取配置文件 .boxci.yml
	confb, err := remote.File(user, repo, build, repo.Config)
	if err != nil {
		logrus.Errorf("error: (%s): cannot find %s in %s: %s", repo.FullName, repo.Config, build.Ref, err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	sha := shasum(confb)
	configStore := s.Manager.ConfigStore()
	conf, err := configStore.ConfigFind(repo, sha)
	if err != nil {
		conf = &models.Config{
			RepoID: repo.ID,
			Data:   string(confb),
			Hash:   sha,
		}
		err = configStore.ConfigCreate(conf)
		if err != nil {
			// retry in case we receive two hooks at the same time
			conf, err = configStore.ConfigFind(repo, sha)
			if err != nil {
				logrus.Errorf("failed to find or persist build config for %s. %s", repo.FullName, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}

	build.ConfigID = conf.ID

	// todo 关于 .netrc 处理

	branches, err := yaml.ParseString(conf.Data)
	if err == nil {
		if !branches.Branches.Match(build.Branch) && build.Event != models.EventTag && build.Event != models.EventDeploy {
			w.WriteHeader(200)
			w.Write([]byte("Branch does not match restrictions defined in yaml."))
			return
		}
	}

	build.RepoID = repo.ID
	build.Verified = true
	build.Status = models.StatusPending

	// todo 一些 build 的限制条件检测

	build.Trim()
	err = s.Manager.Store().CreateBuild(build, build.Procs...)
	if err != nil {
		logrus.Errorf("failed to save commit for %s. %s", repo.FullName, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(200)
	buildB, _ := json.Marshal(build)
	w.Write(buildB)

	if build.Status == models.StatusBlocked {
		return
	}

	envs := map[string]string{}
	if environStore := s.Manager.EnvironStore(); environStore != nil {
		globals, _ := environStore.EnvironList(repo)
		for _, global := range globals {
			envs[global.Name] = global.Value
		}
	}

	secs, err := s.Manager.SecretService().SecretListBuild(repo, build)
	if err != nil {
		logrus.Debugf("Error getting secrets for %s#%s. %s", repo.FullName, build.Number, err)
	}

	regs, err := s.Manager.RegistryService().RegistryList(repo)
	if err != nil {
		logrus.Debugf("Error getting registry credentials fro %s#%d. %s", repo.FullName, build.Number, err)
	}

	// get the previous build so that we can send
	// on status change notifications
	last, _ := s.Manager.Store().GetBuildLastBefore(repo, build.Branch, build.ID)

	defer func() {
		uri := fmt.Sprintf("%s/%s/%d", httplib.GetURL(r), repo.FullName, build.Number)
		err = remote.Status(user, repo, build, uri)
		if err != nil {
			logrus.Errorf("error setting commit status for %s/%d, err:(%s)", repo.FullName, build.Number, err)
		}
	}()

	b := builder{
		Repo: repo,
		Curr: build,
		Last: last,
		Secs: secs,
		Regs: regs,
		Envs: envs,
		Link: httplib.GetURL(r),
		Yaml: conf.Data,
	}
	items, err := b.Build()
	if err != nil {
		build.Status = models.StatusError
		build.Started = time.Now().Unix()
		build.Finished = build.Started
		build.Error = err.Error()
		s.Manager.Store().UpdateBuild(build)
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
					State:   models.StatusPending,
				}
				build.Procs = append(build.Procs, proc)
			}
		}
	}

	err = s.Manager.Store().ProcCreate(build.Procs)
	if err != nil {
		logrus.Errorf("error persisting procs %s/%d: %s", repo.FullName, build.Number, err)
	}

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

type builder struct {
	Repo  *models.Repo
	Curr  *models.Build
	Last  *models.Build
	Netrc *models.Netrc
	Secs  []*models.Secret
	Regs  []*models.Registry
	Link  string
	Yaml  string // 项目的 ci 配置文件
	Envs  map[string]string
}

type buildItem struct {
	Proc     *models.Proc
	Platform string
	Labels   map[string]string
	Config   *backend.Config
}

func (b *builder) Build() ([]*buildItem, error) {
	axes, err := matrix.ParseString(b.Yaml)
	if err != nil {
		return nil, err
	}
	if len(axes) == 0 {
		axes = append(axes, matrix.Axis{})
	}

	var items []*buildItem
	for i, axis := range axes {
		proc := &models.Proc{
			BuildID: b.Curr.ID,
			PID:     i + 1,
			PGID:    i + 1,
			State:   models.StatusPending,
			Environ: axis,
		}

		metadata := metadataFromStruct(b.Repo, b.Curr, b.Last, proc, b.Link)
		environ := metadata.Environ()
		//for k, v := range metadata.EnvironDrone() {
		//	environ[k] = v
		//}
		for k, v := range axis {
			environ[k] = v
		}

		var secrets []compiler.Secret
		for _, sec := range b.Secs {
			if !sec.Match(b.Curr.Event) {
				continue
			}
			secrets = append(secrets, compiler.Secret{
				Name:  sec.Name,
				Value: sec.Value,
				Match: sec.Images,
			})
		}

		y := b.Yaml
		s, err := envsubst.Eval(y, func(name string) string {
			env := environ[name]
			if strings.Contains(env, "\n") {
				env = fmt.Sprintf("%q", env)
			}
			return env
		})
		if err != nil {
			return nil, err
		}
		y = s

		parsed, err := yaml.ParseString(y)
		if err != nil {
			return nil, err
		}
		metadata.Sys.Arch = parsed.Platform
		if metadata.Sys.Arch == "" {
			metadata.Sys.Arch = "linux/amd64"
		}

		lerr := linter.New(
			linter.WithTrusted(b.Repo.IsTrusted),
		).Lint(parsed)
		if lerr != nil {
			return nil, lerr
		}

		var registries []compiler.Registry
		for _, reg := range b.Regs {
			registries = append(registries, compiler.Registry{
				Hostname: reg.Address,
				Username: reg.Username,
				Password: reg.Password,
				Email:    reg.Email,
			})
		}

		ir := compiler.New(
			compiler.WithEnviron(environ),
			compiler.WithEnviron(b.Envs),
			//compiler.WithEscalated(...),
			//compiler.WithResourceLimit(Config.Pipeline.Limits.MemSwapLimit, Config.Pipeline.Limits.MemLimit, Config.Pipeline.Limits.ShmSize, Config.Pipeline.Limits.CPUQuota, Config.Pipeline.Limits.CPUShares, Config.Pipeline.Limits.CPUSet),
			//compiler.WithVolumes(Config.Pipeline.Volumes...),
			//compiler.WithNetworks(Config.Pipeline.Networks...),
			compiler.WithLocal(false),
			//compiler.WithOption(
			//	compiler.WithNetrc(
			//		b.Netrc.Login,
			//		b.Netrc.Password,
			//		b.Netrc.Machine,
			//	),
			//	b.Repo.IsPrivate,
			//),
			compiler.WithRegistry(registries...),
			compiler.WithSecret(secrets...),
			compiler.WithPrefix(
				fmt.Sprintf(
					"%d_%d",
					proc.ID,
					rand.Int(),
				),
			),
			compiler.WithEnviron(proc.Environ),
			compiler.WithProxy(),
			compiler.WithWorkspaceFromURL("/boxci", b.Repo.Link),
			compiler.WithMetadata(metadata),
		).Compile(parsed)

		item := &buildItem{
			Proc:     proc,
			Config:   ir,
			Labels:   parsed.Labels,
			Platform: metadata.Sys.Arch,
		}
		if item.Labels == nil {
			item.Labels = map[string]string{}
		}
		items = append(items, item)
	}
	return items, nil
}

func shasum(raw []byte) string {
	sum := sha256.Sum256(raw)
	return fmt.Sprintf("%x", sum)
}

// return the metadata from the cli context.
func metadataFromStruct(repo *models.Repo, build, last *models.Build, proc *models.Proc, link string) frontend.Metadata {
	return frontend.Metadata{
		Repo: frontend.Repo{
			Name:    repo.FullName,
			Link:    repo.Link,
			Remote:  repo.Clone,
			Private: repo.IsPrivate,
		},
		Curr: frontend.Build{
			Number:   build.Number,
			Parent:   build.Parent,
			Created:  build.Created,
			Started:  build.Started,
			Finished: build.Finished,
			Status:   build.Status,
			Event:    build.Event,
			Link:     build.Link,
			Target:   build.Deploy,
			Commit: frontend.Commit{
				Sha:     build.Commit,
				Ref:     build.Ref,
				Refspec: build.Refspec,
				Branch:  build.Branch,
				Message: build.Message,
				Author: frontend.Author{
					Name:   build.Author,
					Email:  build.Email,
					Avatar: build.Avatar,
				},
			},
		},
		Prev: frontend.Build{
			Number:   last.Number,
			Created:  last.Created,
			Started:  last.Started,
			Finished: last.Finished,
			Status:   last.Status,
			Event:    last.Event,
			Link:     last.Link,
			Target:   last.Deploy,
			Commit: frontend.Commit{
				Sha:     last.Commit,
				Ref:     last.Ref,
				Refspec: last.Refspec,
				Branch:  last.Branch,
				Message: last.Message,
				Author: frontend.Author{
					Name:   last.Author,
					Email:  last.Email,
					Avatar: last.Avatar,
				},
			},
		},
		Job: frontend.Job{
			Number: proc.PID,
			Matrix: proc.Environ,
		},
		Sys: frontend.System{
			Name: "drone",
			Link: link,
			Arch: "linux/amd64",
		},
	}
}
