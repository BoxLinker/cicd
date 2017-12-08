package server

import (
	"net/http"
	"github.com/Sirupsen/logrus"
	"fmt"
	"github.com/BoxLinker/cicd/models"
	"github.com/gorilla/mux"
)

func (s *Server) AuthCodeBase(w http.ResponseWriter, r *http.Request){
	scmType := mux.Vars(r)["scm"]
	logrus.Debugf("SCM Auth (%s)", scmType)
	logrus.Debugf("scmType (%+v) exists (%+v)", scmType, models.SCMExists(scmType))
	if scmType == "" || !models.SCMExists(scmType) {
		http.Error(w, "wrong scm type", http.StatusBadRequest)
		return
	}
	logrus.Debugf("AuthCodeBase ==> %s", scmType)

	scmUser, err := s.Manager.GetSCM(scmType).Authorize(w, r, "boxlinker-cicd")
	if err != nil {
		logrus.Errorf("cannot authenticate user. %s", err)
		http.Redirect(w, r, fmt.Sprintf("%s/?error=oauth_error", s.Config.HomeHost), 301)
		return
	}

	if scmUser == nil {
		return
	}

	uid := s.getCtxUserID(r)
	scmUser.UCenterID = uid

	if u := s.Manager.GetUserByUCenterID(scmUser.UCenterID, scmUser.SCM); u != nil {
		if u.ID <= 0 {
			http.Redirect(w, r, fmt.Sprintf("%s/?error=server_interval_error&err_msg=%s", s.Config.HomeHost, "user id is 0"), 301)
			return
		}
		u.Login = scmUser.Login
		u.Email = scmUser.Email
		u.AccessToken = scmUser.AccessToken
		if err := s.Manager.UpdateUser(u); err != nil {
			http.Redirect(w, r, fmt.Sprintf("%s/?error=server_interval_error&err_msg=%s", s.Config.HomeHost, err.Error()), 301)
			return
		}
	} else {
		if err := s.Manager.SaveUser(scmUser); err != nil {
			http.Redirect(w, r, fmt.Sprintf("%s/?error=server_interval_error&err_msg=%s", s.Config.HomeHost, err.Error()), 301)
			return
		}
	}

	http.Redirect(w, r, fmt.Sprintf("%s/cicd", s.Config.HomeHost), 301)
}

