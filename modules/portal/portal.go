package portal

import (
	"github.com/codegangsta/negroni"
	"net/http"
	"github.com/gorilla/mux"
)

type Engine struct {
	router *mux.Router
}

func New(router *mux.Router) *Engine {
	return &Engine{
		router: router,
	}
}

func (e *Engine) Group(prefix string) *RouterGroup {

	return &RouterGroup{
		engine: e,
		prefix: prefix,
		router: mux.NewRouter().PathPrefix(prefix).Subrouter(),
	}
}


type RouterGroup struct {
	engine *Engine
	prefix string
	router *mux.Router
}

func (g *RouterGroup) Use(handlers ...negroni.Handler) {
	n := negroni.New(handlers...)
	n.UseHandler(g.router)
	g.engine.router.PathPrefix(g.prefix).Handler(n)
}

func (g *RouterGroup) Get(path string, f func(w http.ResponseWriter, r *http.Request)) {
	g.router.HandleFunc(path, f).Methods("GET")
}

func (g *RouterGroup) Post(path string, f func(w http.ResponseWriter, r *http.Request)) {
	g.router.HandleFunc(path, f).Methods("POST")
}

