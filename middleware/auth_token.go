package middleware

import (
	"net/http"
	"github.com/BoxLinker/boxlinker-api"
	"golang.org/x/net/context"
	"github.com/Sirupsen/logrus"
	"github.com/BoxLinker/cicd/auth"
)

type AuthTokenRequired struct {
	authUrl string
	redirectURL string

}

func NewAuthTokenRequired(url string, redirectURL string) Middleware {
	return &AuthTokenRequired{
		authUrl: url,
		redirectURL: redirectURL,
	}
}

type resultAuth struct {
	Status int `json:"status"`
	Results interface{} `json:"results"`
	Msg string `json:"msg"`
}

func (a *AuthTokenRequired) HandlerFuncWithNext(w http.ResponseWriter, r *http.Request, next http.HandlerFunc){
	token := r.Header.Get("X-Access-Token")
	redirect := len(a.redirectURL) > 0
	if token == "" {
		if redirect {
			http.Redirect(w, r, a.redirectURL, 301)
			return
		}
		boxlinker.Resp(w, boxlinker.STATUS_UNAUTHORIZED, nil, "unauthorized")
		return
	}
	logrus.Debugf("AuthToken url: %s", a.authUrl)
	//resp, err := httplib.Get(a.authUrl).Header("X-Access-Token", token).SetTimeout(time.Second*3, time.Second*3).Response()
	//if err != nil {
	//	logrus.Errorf("AuthToken err: %v", err)
	//	boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR,nil)
	//	return
	//}
	//var result resultAuth
	//if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
	//	boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR,nil, err.Error())
	//	return
	//}
	result, err := auth.TokenAuth(a.authUrl, token)
	if err != nil {
		if redirect {
			logrus.Errorf("AuthToken err: %s", err.Error())
			http.Redirect(w, r, a.redirectURL, 301)
			return
		}
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR,nil, err.Error())
		return
	}
	if result.Status == boxlinker.STATUS_OK && next != nil {
		logrus.Debugf("AuthToken result: %+v", result)
		next(w, r.WithContext(context.WithValue(r.Context(), "user", result.Results)))
	} else {
		logrus.Debugf("AuthToken failed: %+v", result)
		if redirect {
			http.Redirect(w, r, a.redirectURL, 301)
			return
		}
		boxlinker.Resp(w, boxlinker.STATUS_UNAUTHORIZED, nil, "unauthorized")
	}
}

