package authn

import (
	"errors"
	"github.com/BoxLinker/boxlinker-api/modules/httplib"
	"time"
	"net/http"
	"fmt"
	"github.com/Sirupsen/logrus"
)

type Labels map[string][]string

type Authenticator interface {
	Authenticate(user string, password PasswordString) (bool, Labels, error)
	Stop()
	Name() string
}

var NoMatch = errors.New("did not match any rule")
var WrongPass = errors.New("wrong pass for user")

type PasswordString string

func (ps PasswordString) String() string {
	if len(ps) == 0 {
		return ""
	}
	return "***"
}

type DefaultAuthenticator struct {
	BasicAuthURL string
}

func (auth *DefaultAuthenticator) Authenticate(user string, password PasswordString)(bool, Labels, error){
	resp, err := httplib.Get(auth.BasicAuthURL).SetBasicAuth(user, string(password)).SetTimeout(time.Second*5, time.Second*10).Response()
	if err != nil {
		logrus.Errorf("basic auth err: ", err)
		return false, nil, err
	}
	logrus.Debugf("basic auth resp: %d", resp.StatusCode)
	if resp.StatusCode == http.StatusOK {
		return true, nil, nil
	} else if resp.StatusCode == http.StatusUnauthorized {
		return false, nil, WrongPass
	} else if resp.StatusCode == http.StatusNotFound {
		return false, nil, NoMatch
	} else {
		return false, nil, errors.New(fmt.Sprintf("auth err: %d", resp.StatusCode))
	}
}

func (auth *DefaultAuthenticator) Stop(){}
func (auth *DefaultAuthenticator) Name() string {
	return "DefaultAuthenticator"
}