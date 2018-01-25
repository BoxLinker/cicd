package authn

import (
	"github.com/BoxLinker/boxlinker-api/modules/httplib"
	"time"
	"github.com/Sirupsen/logrus"
	"encoding/json"
)

type result struct {
	Status int `json:"status"`
	Result struct{
		Token string `json:"token"`
	} `json:"result"`
	Msg string `json:"msg"`
}

type BoxlinkerUCenterAuthenticator struct {
	BasicAuthURL string
}

func (auth *BoxlinkerUCenterAuthenticator) Authenticate(user string, password PasswordString)(bool, Labels, error){
	b, _ := json.Marshal(map[string]string{
		"user_name": user,
		"password": string(password),
	})
	resp, err := httplib.Post(auth.BasicAuthURL).Header("Content-Type", "application/json").
		Body(b).SetTimeout(time.Second*5, time.Second*10).Response()
	if err != nil {
		logrus.Errorf("basic auth err: ", err)
		return false, nil, err
	}
	var res result
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return false, nil, err
	}
	logrus.Debugf("basic auth resp: %+v", res)

	if res.Status == 0 {
		return true, nil, nil
	} else {
		return false, nil, nil
	}
}

func (auth *BoxlinkerUCenterAuthenticator) Stop(){}
func (auth *BoxlinkerUCenterAuthenticator) Name() string {
	return "BoxlinkerUCenterAuthenticator"
}

