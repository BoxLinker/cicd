package auth

import (
	"github.com/cabernety/gopkg/httplib"
	"time"
	"encoding/json"
)

type ResultAuth struct {
	Status int `json:"status"`
	Results interface{} `json:"results"`
	Msg string `json:"msg"`
}

func TokenAuth(authUrl, token string)(*ResultAuth, error){
	resp, err := httplib.Get(authUrl).Header("X-Access-Token", token).SetTimeout(time.Second*3, time.Second*3).Response()
	if err != nil {
		return nil, err
	}
	result := &ResultAuth{}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return nil, err
	}
	return result, nil
}
