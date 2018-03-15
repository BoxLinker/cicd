package middleware

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/BoxLinker/boxlinker-api"
	"github.com/BoxLinker/cicd/auth"
	"github.com/BoxLinker/cicd/store"
	"github.com/Sirupsen/logrus"
	"github.com/cabernety/gopkg/httplib"
	"github.com/codegangsta/negroni"
	"golang.org/x/net/context"
)

type AuthAPITokenRequired struct {
	authUrl string
	store   store.Store
}

func NewAuthAPITokenRequired(url string, s store.Store) negroni.Handler {
	return &AuthAPITokenRequired{
		authUrl: url,
		store:   s,
	}
}

func (a *AuthAPITokenRequired) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	token := r.Header.Get("X-Access-Token")
	scmType := mux.Vars(r)["scm"]
	if token == "" {
		token = httplib.GetQueryParam(r, "access_token")
	}
	// logrus.Debugf("request token (%s)", token)
	if token == "" {
		boxlinker.Resp(w, boxlinker.STATUS_UNAUTHORIZED, nil, "unauthorized")
		return
	}
	// logrus.Debugf("AuthToken url: %s", a.authUrl)
	result, err := auth.TokenAuth(a.authUrl, token)
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}
	if result.Status == boxlinker.STATUS_OK && next != nil {
		// logrus.Debugf("AuthToken result: %+v", result)
		userResults, ok := result.Results.(map[string]interface{})
		if !ok {
			httplib.Resp(w, httplib.STATUS_UNAUTHORIZED, nil, "user results convert err")
			return
		}
		uid := fmt.Sprintf("%s", userResults["uid"])
		if uid == "" {
			httplib.Resp(w, httplib.STATUS_UNAUTHORIZED, nil, "no uid")
			return
		}
		user := a.store.GetUserByUCenterID(uid, scmType)
		if user == nil {
			httplib.Resp(w, httplib.STATUS_UNAUTHORIZED, nil, "cicd user not found")
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), "user", user)))
	} else {
		logrus.Debugf("AuthToken token:(%s) failed: %+v", token, result)
		boxlinker.Resp(w, boxlinker.STATUS_UNAUTHORIZED, nil, "unauthorized")
	}
}
