package server

import (
	"fmt"
	"net/http"

	"github.com/BoxLinker/boxlinker-api"
	"github.com/BoxLinker/cicd/manager"
	"github.com/BoxLinker/cicd/middleware"
	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/negroni"
	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	//"fmt"

	"github.com/BoxLinker/cicd/models"
)

type Config struct {
	TokenAuthURL string
	HomeHost     string
	RepoConfig   string
	PipeLine     struct {
		Limits     models.ResourceLimit
		Volumes    []string
		Networks   []string
		Privileged []string
	}
}

type Server struct {
	Listen  string
	Manager *manager.DefaultManager
	Config  Config
}

func (s *Server) Run() error {

	logrus.Debugf("Server Config: \n%+v", s.Config)

	loginRequired := middleware.NewAuthAPITokenRequired(s.Config.TokenAuthURL, s.Manager.Store())
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

	router.HandleFunc("/v1/cicd/{scm}/hook", s.Hook).Methods("POST")

	rRouter := getRouter(router, "/v1/cicd/user/scms", loginRequired)
	{
		rRouter.HandleFunc("", s.GetScms).Methods("GET")
	}

	userRouter := getRouter(router, "/v1/cicd/{scm}/user",
		scmRequired, loginRequired)
	{
		userRouter.HandleFunc("/repos", s.GetRepos).Methods("GET")
	}

	repoRouter := getRouter(router, "/v1/cicd/{scm}/repos/{owner}/{name}",
		scmRequired, loginRequired, setRepoM)
	{
		repoRouter.HandleFunc("", s.PostRepo).Methods("POST")
		repoRouter.HandleFunc("", s.GetRepo).Methods("GET")
		repoRouter.HandleFunc("/procs/{build_id}", s.GetProcs).Methods("GET")
		repoRouter.HandleFunc("/logs/{number}/{pid}", s.GetProcLogs).Methods("GET")
		repoRouter.HandleFunc("/builds/{number}", s.PostBuild).Methods("POST")
		repoRouter.HandleFunc("/builds/{number}", s.GetBuild).Methods("GET")
		repoRouter.HandleFunc("/builds", s.QueryBuild).Methods("GET")
		repoRouter.HandleFunc("/branches", s.GetRepoBranches).Methods("GET")
		repoRouter.HandleFunc("/query_branch_build", s.QueryRepoBranchBuilding).Methods("GET")
		repoRouter.HandleFunc("/search_build", s.SearchRepoBuilding).Methods("GET")
	}

	streamRouter := getRouter(router, "/v1/cicd/{scm}/stream/logs/{owner}/{name}", scmRequired, loginRequired, setRepoM)
	{
		streamRouter.HandleFunc("/{build}/{number}", s.LogStream).Methods("GET")
	}

	authorizeRouter := getRouter(router, "/v1/cicd/{scm}",
		authorizeTokenM)
	{
		authorizeRouter.HandleFunc("/authorize", s.AuthCodeBase).Methods("POST", "GET")
	}

	ss := http.Server{
		Addr:    s.Listen,
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
	userData := a.getUserInfo(r)
	if userData != nil {
		return userData.UCenterID
	}
	return ""
	// us := r.Context().Value("user")
	// if us == nil {
	// 	return ""
	// }
	// ctx := us.(map[string]interface{})
	// if ctx == nil || ctx["uid"] == nil {
	// 	return ""
	// }
	// return ctx["uid"].(string)
}

func (a *Server) getUserInfo(r *http.Request) *models.User {
	// scm := mux.Vars(r)["scm"]
	userData, ok := r.Context().Value("user").(*models.User)
	if !ok {
		return nil
	}
	return userData
	// return a.Manager.Store().GetUserByUCenterID(a.getCtxUserID(r), scm)
}
