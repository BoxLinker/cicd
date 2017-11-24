package middleware

import (
	"net/http"
	"github.com/BoxLinker/boxlinker-api"
	"golang.org/x/net/context"
	"github.com/Sirupsen/logrus"
	"github.com/BoxLinker/cicd/auth"
	"github.com/cabernety/gopkg/httplib"
)

type AuthorizeTokenRequired struct {
	authUrl string
	redirectURL string

}

func NewAuthorizeTokenRequired(url string, redirectURL string) Middleware {
	return &AuthorizeTokenRequired{
		authUrl: url,
		redirectURL: redirectURL,
	}
}

func (a *AuthorizeTokenRequired) HandlerFuncWithNext(w http.ResponseWriter, r *http.Request, next http.HandlerFunc){
	token := httplib.GetCookie(r, "X-Access-Token")
	logrus.Debugf("request token (%s)", token)
	if token == "" {
		http.Redirect(w, r, a.redirectURL, 301)
		return
	}
	logrus.Debugf("AuthToken url: %s", a.authUrl)
	result, err := auth.TokenAuth(a.authUrl, token)
	if err != nil {
		logrus.Errorf("AuthToken err: %s", err.Error())
		http.Redirect(w, r, a.redirectURL, 301)
		return
	}
	if result.Status == boxlinker.STATUS_OK && next != nil {
		logrus.Debugf("AuthToken result: %+v", result)
		next(w, r.WithContext(context.WithValue(r.Context(), "user", result.Results)))
	} else {
		logrus.Debugf("AuthToken failed: %+v", result)
		http.Redirect(w, r, a.redirectURL, 301)
		return
	}
}

