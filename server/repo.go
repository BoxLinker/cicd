package server

import (
	"net/http"
	"github.com/BoxLinker/boxlinker-api"
	"strconv"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/BoxLinker/cicd/models"
	"fmt"
	"encoding/base32"
	"github.com/gorilla/securecookie"
	"github.com/BoxLinker/cicd/modules/token"
	"github.com/cabernety/gopkg/httplib"
)

func (s *Server) GetRepos(w http.ResponseWriter, r *http.Request) {
	flush, _ := strconv.ParseBool(boxlinker.GetQueryParam(r, "flush"))
	pc := boxlinker.ParsePageConfig(r)
	u := s.getUserInfo(r)
	logrus.Debugf("GetRepos (%s)", u.SCM)

	if flush {
		if repos, err := s.Manager.GetSCM(u.SCM).Repos(u); err != nil {
			boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
			return
		} else {
			if err := s.Manager.RepoBatch(u, repos); err != nil {
				boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
				return
			}
		}
	}

	boxlinker.Resp(w, boxlinker.STATUS_OK, pc.PaginationResult(s.Manager.QueryRepos(u, &pc)))
}

func (s *Server) PostRepo(w http.ResponseWriter, r *http.Request) {
	scmType := boxlinker.GetQueryParam(r, "scm")
	if scmType == "" || !models.SCMExists(scmType) {
		http.Error(w, "wrong scm type", http.StatusBadRequest)
		return
	}
	remote := s.Manager.GetSCM(scmType)
	_ = remote
	user := s.getUserInfo(r)
	owner := mux.Vars(r)["owner"]
	repoName := mux.Vars(r)["name"]
	logrus.Debugf("PostRepo remote(%s) user(%s) owner(%s) repo(%s)", scmType, user.Login, owner, repoName)
	repo, err := s.Manager.GetRepoOwnerName(owner, repoName)
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, fmt.Sprintf("repo (%s/%s) not found: %s", owner, repoName, err.Error()))
		return
	}

	if repo.IsActive {
		boxlinker.Resp(w, 409, nil, "Repository is already active.")
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
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}

	link := fmt.Sprintf(
		"%s/v1/cicd/hook/{%s}?access_token=%s",
		repo.SCM,
		httplib.GetURL(r),
		sig,
	)

	err = remote.Activate(user, repo, link)
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}

	from, err := remote.Repo(user, repo.Owner, repo.Name)
	if err == nil {
		repo.Update(from)
	}

	err = s.Manager.Store().UpdateRepo(repo)
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}

	boxlinker.Resp(w, boxlinker.STATUS_OK, repo)
}