package server

import (
	"github.com/BoxLinker/boxlinker-api"
	"github.com/gorilla/mux"
	"net/http"
	"github.com/gorilla/context"
	"github.com/BoxLinker/cicd/codebase"
	"github.com/BoxLinker/cicd/middleware"
	"github.com/codegangsta/negroni"
	"github.com/Sirupsen/logrus"
	"github.com/BoxLinker/cicd/manager"
)

type Config struct {
	TokenAuthURL string
	HomeHost string
}

type Server struct {
	CodeBase codebase.CodeBase
	Listen string
	Manager *manager.DefaultManager
	Config Config
}

func (s *Server) Run() error {

	logrus.Debugf("Server Config: \n%+v", s.Config)
	cs := boxlinker.Cors

	globalMux := http.NewServeMux()

	apiRouter := mux.NewRouter()

	tokenAuthRequired := middleware.NewAuthTokenRequired(s.Config.TokenAuthURL)
	tokenAuthRouter := negroni.New()
	tokenAuthRouter.Use(negroni.HandlerFunc(tokenAuthRequired.HandlerFuncWithNext))
	tokenAuthRouter.UseHandler(apiRouter)
	globalMux.Handle("/v1/cicd/auth", tokenAuthRouter)

	// git callback
	authorizeRouter := mux.NewRouter()
	authorizeRouter.HandleFunc("/v1/cicd/authorize", s.AuthCodeBase).Methods("GET", "POST")
	globalMux.Handle("/v1/cicd/authorize", authorizeRouter)

	ss := http.Server{
		Addr: s.Listen,
		Handler: context.ClearHandler(cs.Handler(globalMux)),
	}

	logrus.Infof("CICD server running on %s", s.Listen)
	return ss.ListenAndServe()
}