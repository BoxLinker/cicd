package server

import (
	"github.com/BoxLinker/boxlinker-api"
	"github.com/gorilla/mux"
	"net/http"
	"github.com/gorilla/context"
	"github.com/BoxLinker/cicd/middleware"
	"github.com/codegangsta/negroni"
	"github.com/Sirupsen/logrus"
	"github.com/BoxLinker/cicd/manager"
	//"fmt"
	"github.com/BoxLinker/cicd/models"
	"fmt"
)

type Config struct {
	TokenAuthURL string
	HomeHost string
}

type Server struct {
	Listen   string
	Manager  *manager.DefaultManager
	Config   Config
}

func (s *Server) Run() error {

	logrus.Debugf("Server Config: \n%+v", s.Config)
	cs := boxlinker.Cors
	globalMux := http.NewServeMux()

	apiTokenM := middleware.NewAuthAPITokenRequired(s.Config.TokenAuthURL)
	scmRequired := middleware.NewSCMRequired()
	authorizeTokenM := middleware.NewAuthorizeTokenRequired(s.Config.TokenAuthURL, fmt.Sprintf("%s/login", s.Config.HomeHost))


	apiRouter := mux.NewRouter()
	apiRouter.HandleFunc("/v1/cicd/auth/repos", s.GetRepos).Methods("GET")

	authRouter := negroni.New()
	authRouter.Use(negroni.HandlerFunc(scmRequired.HandlerFuncWithNext))
	authRouter.Use(negroni.HandlerFunc(apiTokenM.HandlerFuncWithNext))
	authRouter.UseHandler(apiRouter)
	globalMux.Handle("/v1/cicd/auth/", authRouter)

	// codebase auth router
	authorizeRouter := mux.NewRouter()
	authorizeRouter.HandleFunc("/v1/cicd/authorize/{scm}", s.AuthCodeBase).Methods("GET", "POST")
	tokenAuthRedirectRouter := negroni.New()
	tokenAuthRedirectRouter.Use(negroni.HandlerFunc(authorizeTokenM.HandlerFuncWithNext))
	tokenAuthRedirectRouter.UseHandler(authorizeRouter)
	globalMux.Handle("/v1/cicd/authorize/", tokenAuthRedirectRouter)

	ss := http.Server{
		Addr: s.Listen,
		Handler: context.ClearHandler(cs.Handler(globalMux)),
	}

	logrus.Infof("CICD server running on %s", s.Listen)
	return ss.ListenAndServe()
}

func (a *Server) getCtxUserID(r *http.Request) string {
	us := r.Context().Value("user")
	if us == nil {
		return ""
	}
	ctx := us.(map[string]interface{})
	if ctx == nil || ctx["uid"] == nil {
		return ""
	}
	return ctx["uid"].(string)
}

func (a *Server) getUserInfo(r *http.Request) *models.SCMUser {
	scm := boxlinker.GetQueryParam(r, "scm")
	return a.Manager.GetSCMUserByUCenterID(a.getCtxUserID(r), scm)
}