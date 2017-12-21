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
	RepoConfig string
	PipeLine struct{
		Limits models.ResourceLimit
		Volumes []string
		Networks []string
		Privileged []string
	}
}

type Server struct {
	Listen   string
	Manager  *manager.DefaultManager
	Config   Config
}

func (s *Server) Run() error {

	logrus.Debugf("Server Config: \n%+v", s.Config)

	loginRequired := middleware.NewAuthAPITokenRequired(s.Config.TokenAuthURL)
	scmRequired := middleware.NewSCMRequired()
	authorizeTokenM := middleware.NewAuthorizeTokenRequired(s.Config.TokenAuthURL, fmt.Sprintf("%s/login", s.Config.HomeHost))
	setRepoM := middleware.NewSetRepo(s.Manager)

	//globalMux := http.NewServeMux()

	router := mux.NewRouter()

	//hookRouter := mux.NewRouter()
	//hookRouter.HandleFunc("/v1/cicd/hook/{scm}", s.Hook).Methods("POST")
	//globalMux.Handle("/v1/cicd/hook/", hookRouter)
	//
	//userRouter := mux.NewRouter()
	//userRouter.HandleFunc("/v1/cicd/user/repos", s.GetRepos).Methods("GET")
	//userRouterN := negroni.New()
	//userRouterN.UseHandler(userRouter)
	//globalMux.Handle("/v1/cicd/user", userRouterN)

	router.HandleFunc("/v1/cicd/hook/{scm}", s.Hook).Methods("POST")

	userRouter := getRouter(router, "/v1/cicd/user",
		loginRequired, scmRequired)
	{
		userRouter.HandleFunc("/repos", s.GetRepos).Methods("GET")
	}

	repoRouter := getRouter(router, "/v1/cicd/repos/{owner}/{name}",
		loginRequired, scmRequired, setRepoM)
	{
		repoRouter.HandleFunc("/logs/{number}/{pid}", s.GetProcLogs).Methods("GET")
		repoRouter.HandleFunc("", s.PostRepo).Methods("POST")
	}

	authorizeRouter := getRouter(router, "/v1/cicd/authorize",
		authorizeTokenM)
	{
		authorizeRouter.HandleFunc("/{scm}", s.AuthCodeBase).Methods("POST", "GET")
	}

	//repoRouter := mux.NewRouter().
	//	PathPrefix("/v1/cicd/repos/{owner}/{name}").Subrouter()
	//repoRouter.HandleFunc("/logs/{number}/{pid}", s.GetProcLogs).Methods("GET")
	//repoRouter.HandleFunc("", s.PostRepo).Methods("POST")
	//
	//router.PathPrefix("/v1/cicd/repos/{owner}/{name}").
	//	Handler(negroni.New(
	//		loginRequired,
	//		scmRequired,
	//		setRepoM,
	//		negroni.Wrap(repoRouter),
	//	))

	//authorizeRouter := mux.NewRouter()
	//authorizeRouter.HandleFunc("/v1/cicd/authorize/{scm}", s.AuthCodeBase).Methods("GET", "POST")
	//tokenAuthRedirectRouter := negroni.New(authorizeTokenM)
	//tokenAuthRedirectRouter.UseHandler(authorizeRouter)
	//globalMux.Handle("/v1/cicd/authorize/", tokenAuthRedirectRouter)


	ss := http.Server{
		Addr: s.Listen,
		Handler: context.ClearHandler(boxlinker.Cors.Handler(router)),
	}

	logrus.Infof("CICD server running on %s", s.Listen)
	return ss.ListenAndServe()
}

func getRouter(pRouter *mux.Router, path string, middlewares ...negroni.Handler) *mux.Router {
	subRouter := mux.NewRouter().PathPrefix(path).Subrouter()
	r := negroni.New(middlewares...)
	r.UseHandler(subRouter)
	pRouter.PathPrefix(path).Handler(r)
	return subRouter
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

func (a *Server) getUserInfo(r *http.Request) *models.User {
	scm := boxlinker.GetQueryParam(r, "scm")
	return a.Manager.GetUserByUCenterID(a.getCtxUserID(r), scm)
}