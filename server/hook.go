package server

import (
	"net/http"
	"github.com/gorilla/mux"
	"github.com/BoxLinker/cicd/models"
	"github.com/Sirupsen/logrus"
	"regexp"
)

var skipRe = regexp.MustCompile(`\[(?i:ci *skip|skip *ci)\]`)

func (s *Server) Hook(w http.ResponseWriter, r *http.Request) {
	scmType := mux.Vars(r)["scm"]
	if !models.SCMExists(scmType) {
		http.Error(w,  "bad scm type", http.StatusBadRequest)
		return
	}

	remote := s.Manager.GetSCM(scmType)
	tmpRepo, build, err := remote.Hook(r)
	if err != nil {
		logrus.Errorf("failure to parse hook. %s", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	if build == nil {
		w.WriteHeader(200)
		return
	}
	if tmpRepo == nil {
		logrus.Errorf("failure to ascertain repo from hook.")
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	// 如果 commit message 里的信息有类似 skip ci 等字样，那就忽略此次提交
	skipMatch := skipRe.FindString(build.Message)
	if len(skipMatch) > 0 {
		logrus.Infof("ignoring hook. %s found in %s.", skipMatch, build.Message)
		w.WriteHeader(204)
		return
	}

	repo, err := s.Manager.GetRepoOwnerName(tmpRepo.Owner, tmpRepo.Name)
	if err != nil {
		logrus.Errorf("failed to find repo %s/%s from hook. %s", tmpRepo.Owner, tmpRepo.Name, err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// todo check whether repo is active

	// todo get the token and verify the hook is authroized

	if repo.UserID == 0 {
		logrus.Warnf("ignoring hook. repo %s has no owner.", repo.FullName)
		w.WriteHeader(204)
		return
	}

	var skipped = true
	if (build.Event == models.EventPush && repo.AllowPush) ||
		(build.Event == models.EventPull && repo.AllowPull) ||
		(build.Event == models.EventDeploy && repo.AllowDeploy) ||
		(build.Event == models.EventTag && repo.AllowTag) {
			skipped = false
	}

	if skipped {
		logrus.Infof("ignoring hook. repo %s is disabled.", repo.FullName)
		w.WriteHeader(204)
		return
	}



}
