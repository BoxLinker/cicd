package user

import (
	"net/http"
	"github.com/BoxLinker/boxlinker-api"
	"github.com/Sirupsen/logrus"
)

func (a *Api) AuthToken(w http.ResponseWriter, r *http.Request) {
	u := r.Context().Value("user")
	logrus.Debugf("AuthToken result: %+v", u)
	boxlinker.Resp(w, boxlinker.STATUS_OK, u)
}

