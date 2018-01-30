package middleware

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/BoxLinker/boxlinker-api"
	"github.com/BoxLinker/cicd/models"
	"github.com/codegangsta/negroni"
)

type SCMRequired struct {
}

func NewSCMRequired() negroni.Handler {
	return &SCMRequired{}
}

func (a *SCMRequired) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	s := mux.Vars(r)["scm"]
	if s == "" || !models.SCMExists(s) {
		boxlinker.Resp(w, 400, nil, fmt.Sprintf("scm(%s) param err", s))
		return
	}

	next(w, r)
}
