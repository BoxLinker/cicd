package middleware

import (
	"net/http"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/mux"
	"github.com/BoxLinker/cicd/manager"
	"github.com/BoxLinker/boxlinker-api"
	"context"
	"github.com/Sirupsen/logrus"
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
	logrus.Debugf("==> %s", r.URL.Path)
	logrus.Debugf("[Middleware] SetRepo: owner(%s) name(%s)", owner, name)
	repo, err := s.manager.Store().GetRepoOwnerName(owner, name)
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, nil, err.Error())
		return
	}
	next(w, r.WithContext(context.WithValue(r.Context(), "repo", repo)))
}