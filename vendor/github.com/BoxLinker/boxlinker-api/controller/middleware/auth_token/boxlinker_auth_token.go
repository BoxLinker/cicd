package auth_token

import (
	"net/http"
	"github.com/BoxLinker/boxlinker-api"
	"github.com/Sirupsen/logrus"
	"github.com/BoxLinker/boxlinker-api/modules/httplib"
	"time"
	"encoding/json"
	"golang.org/x/net/context"
)

type BoxlinkerAuthTokenRequired struct {
	authUrl string
}

func NewBoxlinkerAuthTokenRequired(url string) *BoxlinkerAuthTokenRequired {
	return &BoxlinkerAuthTokenRequired{
		authUrl: url,
	}
}

type boxlinkerResultAuth struct {
	Status int `json:"status"`
	Result struct{
		UserUUID string `json:"user_uuid"`
		Username string `json:"user_name"`
	} `json:"result"`
	Msg string `json:"msg"`
}

func (a *BoxlinkerAuthTokenRequired) HandlerFuncWithNext(w http.ResponseWriter, r *http.Request, next http.HandlerFunc){
	tokenHeaderName := "token"
	token := r.Header.Get(tokenHeaderName)
	if token == "" {
		boxlinker.Resp(w, boxlinker.STATUS_UNAUTHORIZED, nil)
		return
	}
	logrus.Debugf("AuthToken url: %s, token: %s", a.authUrl, token)
	resp, err := httplib.Get(a.authUrl).Header(tokenHeaderName, token).SetTimeout(time.Second*3, time.Second*3).Response()
	if err != nil {
		logrus.Errorf("AuthToken err: %v", err)
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR,nil)
		return
	}
	var result boxlinkerResultAuth
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR,nil, err.Error())
		return
	}
	if result.Status == boxlinker.STATUS_OK && next != nil {
		logrus.Debugf("AuthToken result: %+v", result)
		next(w, r.WithContext(context.WithValue(r.Context(), "user", map[string]interface{}{
			"uid": result.Result.UserUUID,
			"username": result.Result.Username,
		})))
	} else {
		logrus.Debugf("AuthToken failed: %+v", result)
		boxlinker.Resp(w, boxlinker.STATUS_UNAUTHORIZED, nil)
	}
}



