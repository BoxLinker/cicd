package middleware

import (
	"context"
	"net/http"

	"github.com/BoxLinker/boxlinker-api"
	"github.com/BoxLinker/cicd/manager"
	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
)

type setRepo struct {
	manager *manager.DefaultManager
}

func NewSetRepo(manager *manager.DefaultManager) negroni.Handler {
	return &setRepo{
		manager: manager,
	}
}

func (s *setRepo) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	owner := mux.Vars(r)["owner"]
	name := mux.Vars(r)["name"]
	logrus.Debugf("[Middleware] SetRepo: owner(%s) name(%s)", owner, name)
	repo, err := s.manager.Store().GetRepoOwnerName(owner, name)
	if err != nil {
		logrus.Errorf("GetRepo err: %v", err)
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, nil, err.Error())
		return
	}
	logrus.Debugf("[Middleware] SetRepo: id(%d) fullName(%s)", repo.ID, repo.FullName)
	next(w, r.WithContext(context.WithValue(r.Context(), "repo", repo)))
}
