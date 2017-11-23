package server

import (
	"net/http"
	"github.com/Sirupsen/logrus"
	"fmt"
	"github.com/BoxLinker/boxlinker-api"
	"github.com/BoxLinker/cicd/models"
)

func (s *Server) AuthCodeBase(w http.ResponseWriter, r *http.Request){
	scmType := models.SCMType(boxlinker.GetQueryParam(r, "scm"))
	if scmType == "" || !scmType.Exists() {
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

	if u := s.Manager.GetSCMUserByUCenterID(scmUser.UCenterID, scmUser.SCM); u != nil {
		if err := s.Manager.UpdateSCMUser(scmUser); err != nil {
			http.Redirect(w, r, fmt.Sprintf("%s/?error=server_interval_error&err_msg=%s", s.Config.HomeHost, err.Error()), 301)
			return
		}
	} else {
		if err := s.Manager.SaveSCMUser(scmUser); err != nil {
			http.Redirect(w, r, fmt.Sprintf("%s/?error=server_interval_error&err_msg=%s", s.Config.HomeHost, err.Error()), 301)
			return
		}
	}

	http.Redirect(w, r, fmt.Sprintf("%s/cicd", s.Config.HomeHost), 301)
}

