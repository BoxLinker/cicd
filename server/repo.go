package server

import (
	"net/http"
	"github.com/BoxLinker/boxlinker-api"
	"strconv"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/BoxLinker/cicd/models"
	"fmt"
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
	logrus.Debugf("PostRepo remote(%s) user(%s) owner(%s) repo(%s)", scmType, user.ID, owner, repoName)
	repo, err := s.Manager.GetRepoOwnerName(owner, repoName)
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, fmt.Sprintf("repo (%s/%s) not found: %s", owner, repoName, err.Error()))
		return
	}

	boxlinker.Resp(w, boxlinker.STATUS_OK, repo)
}