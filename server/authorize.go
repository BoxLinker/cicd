package server

import (
	"net/http"
	"github.com/Sirupsen/logrus"
	"fmt"
	"github.com/BoxLinker/cicd/auth"
	"github.com/cabernety/gopkg/httplib"
	"github.com/BoxLinker/boxlinker-api"
)

func (s *Server) AuthCodeBase(w http.ResponseWriter, r *http.Request){
	loginUrl := fmt.Sprintf("%s/login", s.Config.HomeHost)
	token := httplib.GetCookie(r, "X-Access-Token")
	if len(token) == 0 {
		http.Redirect(w, r, loginUrl, 301)
		return
	}
	codebaseUser, err := s.CodeBase.Authorize(w, r, token)
	if err != nil {
		logrus.Errorf("cannot authenticate user. %s", err)
		http.Redirect(w, r, fmt.Sprintf("%s/?error=oauth_error", s.Config.HomeHost), 301)
		return
	}

	if codebaseUser == nil {
		return
	}

	result, err := auth.TokenAuth(s.Config.TokenAuthURL, codebaseUser.Token)
	if err != nil {
		http.Redirect(w, r, loginUrl, 301)
		return
	}
	if result.Status != boxlinker.STATUS_OK {
		http.Redirect(w, r, loginUrl, 301)
		return
	}

	uid := result.Results.(map[string]interface{})["uid"].(string)
	codebaseUser.UserID = uid

	if has, _ := s.Manager.IsCodeBaseUserExists(codebaseUser.UserID, codebaseUser.Kind); has {
		if err := s.Manager.UpdateCodeBaseUser(codebaseUser); err != nil {
			http.Redirect(w, r, fmt.Sprintf("%s/?error=server_interval_error&err_msg=%s", s.Config.HomeHost, err.Error()), 301)
			return
		}
	} else {
		if err := s.Manager.SaveCodeBaseUser(codebaseUser); err != nil {
			http.Redirect(w, r, fmt.Sprintf("%s/?error=server_interval_error&err_msg=%s", s.Config.HomeHost, err.Error()), 301)
			return
		}
	}

	http.Redirect(w, r, fmt.Sprintf("%s/cicd", s.Config.HomeHost), 301)
}

