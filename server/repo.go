package server

import (
	"encoding/base32"
	"fmt"
	"net/http"
	"strconv"

	"github.com/BoxLinker/cicd/models"
	"github.com/BoxLinker/cicd/modules/token"
	"github.com/Sirupsen/logrus"
	"github.com/cabernety/gopkg/httplib"
	"github.com/gorilla/mux"
	"github.com/gorilla/securecookie"
)

func (s *Server) GetRepos(w http.ResponseWriter, r *http.Request) {
	flush, _ := strconv.ParseBool(httplib.GetQueryParam(r, "flush"))
	pc := httplib.ParsePageConfig(r)
	u := s.getUserInfo(r)
	logrus.Debugf("GetRepos (%s)", u.SCM)

	if flush {
		if repos, err := s.Manager.GetSCM(u.SCM).Repos(u); err != nil {
			httplib.Resp(w, httplib.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
			return
		} else {
			if err := s.Manager.RepoBatch(u, repos); err != nil {
				httplib.Resp(w, httplib.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
				return
			}
		}
	}

	httplib.Resp(w, httplib.STATUS_OK, pc.PaginationResult(s.Manager.QueryRepos(u, &pc)))
}

func (s *Server) PostRepo(w http.ResponseWriter, r *http.Request) {
	scmType := mux.Vars(r)["scm"]
	remote := s.Manager.GetSCM(scmType)
	_ = remote
	user := s.getUserInfo(r)
	owner := mux.Vars(r)["owner"]
	repoName := mux.Vars(r)["name"]
	logrus.Debugf("PostRepo remote(%s) user(%s) owner(%s) repo(%s)", scmType, user.Login, owner, repoName)
	repo, err := s.Manager.GetRepoOwnerName(owner, repoName, scmType)
	if err != nil {
		httplib.Resp(w, httplib.STATUS_NOT_FOUND, fmt.Sprintf("repo (%s/%s) not found: %s", owner, repoName, err.Error()))
		return
	}

	if repo.IsActive {
		httplib.Resp(w, 409, nil, "Repository is already active.")
		return
	}

	repo.IsActive = true
	repo.UserID = user.ID
	if !repo.AllowPush && !repo.AllowPull && !repo.AllowDeploy && !repo.AllowTag {
		repo.AllowPush = true
		repo.AllowPull = true
	}

	if repo.Visibility == "" {
		repo.Visibility = models.VisibilityPublic
		if repo.IsPrivate {
			repo.Visibility = models.VisibilityPrivate
		}
	}

	if repo.Config == "" {
		repo.Config = s.Config.RepoConfig
	}

	if repo.Timeout == 0 {
		repo.Timeout = 60 // 1 hour default build time
	}
	if repo.Hash == "" {
		repo.Hash = base32.StdEncoding.EncodeToString(
			securecookie.GenerateRandomKey(32),
		)
	}

	t := token.New(token.HookToken, repo.FullName)
	sig, err := t.Sign(repo.Hash)
	if err != nil {
		httplib.Resp(w, httplib.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}

	link := fmt.Sprintf(
		"%s/v1/cicd/hook/%s?access_token=%s",
		httplib.GetURL(r),
		repo.SCM,
		sig,
	)

	err = remote.Activate(user, repo, link)
	if err != nil {
		httplib.Resp(w, httplib.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}

	from, err := remote.Repo(user, repo.Owner, repo.Name)
	if err == nil {
		repo.Update(from)
	}

	err = s.Manager.Store().UpdateRepo(repo)
	if err != nil {
		httplib.Resp(w, httplib.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}

	httplib.Resp(w, httplib.STATUS_OK, repo)
}

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

// QueryRepoBranchBuilding 查询指定分支下的最近 5 条构建记录
func (s *Server) QueryRepoBranchBuilding(w http.ResponseWriter, r *http.Request) {
	repo := r.Context().Value("repo").(*models.Repo)
	branch := httplib.GetQueryParam(r, "branch")
	builds := s.Manager.Store().QueryBranchBuild(repo, branch)
	httplib.Resp(w, httplib.STATUS_OK, builds)
}

/*
SearchRepoBuilding 获取 repo 的构建记录, 默认按时间倒序
 params:
	search 搜索, 可以为 分支、用户名 等 ，如果指定 search 则 branch 参数忽略
	branch 指定分支
	currentPage
	pageCount
*/
func (s *Server) SearchRepoBuilding(w http.ResponseWriter, r *http.Request) {
	repo := r.Context().Value("repo").(*models.Repo)
	pager := httplib.ParsePageConfig(r)
	search := httplib.GetQueryParam(r, "search")
	builds := s.Manager.Store().SearchBuild(repo, search, &pager)
	httplib.Resp(w, httplib.STATUS_OK, pager.FormatOutput(builds))
}

// GetRepoBranches 根据 repo 获取 repo 的分支信息
func (s *Server) GetRepoBranches(w http.ResponseWriter, r *http.Request) {
	repo := r.Context().Value("repo").(*models.Repo)
	scmType := mux.Vars(r)["scm"]
	refresh := httplib.GetQueryParam(r, "refresh")
	user := s.getUserInfo(r)
	pager := httplib.ParsePageConfig(r)
	remote := s.Manager.GetSCM(scmType)
	if refresh == "1" {
		branches, err := remote.Branches(user, repo.Owner, repo.Name)
		if err != nil {
			httplib.Resp(w, httplib.STATUS_NOT_FOUND, nil, fmt.Sprintf("branches not found: %v", err))
			return
		}
		if err := s.Manager.Store().BranchBatch(repo, branches); err != nil {
			httplib.Resp(w, httplib.STATUS_INTERNAL_SERVER_ERR, nil, fmt.Sprintf("save branches err: %v", err))
			return
		}
	}
	branches := s.Manager.Store().BranchList(repo, pager.Limit(), pager.Offset())
	pager.TotalCount = len(branches)

	for _, branch := range branches {
		branch.RepoID = repo.ID
	}
	httplib.Resp(w, httplib.STATUS_OK, pager.FormatOutput(branches))
}
