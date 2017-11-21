package server

import (
	"k8s.io/client-go/kubernetes"
	"github.com/go-xorm/xorm"
	"github.com/BoxLinker/boxlinker-api"
	"github.com/gorilla/mux"
	"net/http"
	"github.com/gorilla/context"
	"github.com/BoxLinker/cicd/codebase"
	"github.com/BoxLinker/cicd/middleware"
	"github.com/codegangsta/negroni"
	"github.com/urfave/cli"
)

type Server struct {
	CodeBase codebase.CodeBase
	Listen string
	ClientSet *kubernetes.Clientset
	DBEngine *xorm.Engine
}

func (s *Server) Run(c *cli.Context) error {
	cs := boxlinker.Cors

	globalMux := http.NewServeMux()

	apiRouter := mux.NewRouter()
	apiRouter.HandleFunc("/v1/cicd/authorize", s.AuthCodeBase).Methods("GET", "POST")

	tokenAuthRequired := middleware.NewAuthTokenRequired(c.String("token-auth-url"))
	tokenAuthRouter := negroni.New()
	tokenAuthRouter.Use(negroni.HandlerFunc(tokenAuthRequired.HandlerFuncWithNext))
	tokenAuthRouter.UseHandler(apiRouter)
	globalMux.Handle("/v1/cicd/", tokenAuthRouter)

	ss := http.Server{
		Addr: s.Listen,
		Handler: context.ClearHandler(cs.Handler(globalMux)),
	}

	return ss.ListenAndServe()
}