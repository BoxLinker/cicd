package middleware

import (
	"net/http"
	"github.com/BoxLinker/boxlinker-api"
	"fmt"
	"github.com/BoxLinker/cicd/models"
	"github.com/codegangsta/negroni"
)

type SCMRequired struct {
}

func NewSCMRequired() negroni.Handler {
	return &SCMRequired{
	}
}


func (a *SCMRequired) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc){
	s := boxlinker.GetQueryParam(r, "scm")
	if s == "" || !models.SCMExists(s) {
		boxlinker.Resp(w, boxlinker.STATUS_PARAM_ERR, nil, fmt.Sprintf("scm(%s) param err", s))
		return
	}
	next(w, r)
}

