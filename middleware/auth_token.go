package middleware

import (
	"net/http"

	"github.com/BoxLinker/boxlinker-api"
	"github.com/BoxLinker/cicd/auth"
	"github.com/Sirupsen/logrus"
	"github.com/cabernety/gopkg/httplib"
	"github.com/codegangsta/negroni"
	"golang.org/x/net/context"
)

type AuthAPITokenRequired struct {
	authUrl string
}

func NewAuthAPITokenRequired(url string) negroni.Handler {
	return &AuthAPITokenRequired{
		authUrl: url,
	}
}

func (a *AuthAPITokenRequired) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	token := r.Header.Get("X-Access-Token")
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
		next(w, r.WithContext(context.WithValue(r.Context(), "user", result.Results)))
	} else {
		logrus.Debugf("AuthToken token:(%s) failed: %+v", token, result)
		boxlinker.Resp(w, boxlinker.STATUS_UNAUTHORIZED, nil, "unauthorized")
	}
}
