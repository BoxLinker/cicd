package server

import (
	"net/http"
	"github.com/BoxLinker/cicd/models"
	"github.com/gorilla/mux"
	"strconv"
	"github.com/BoxLinker/boxlinker-api"
	"fmt"
	"io"
	"github.com/Sirupsen/logrus"
)

func (s *Server) GetProcLogs(w http.ResponseWriter, r *http.Request) {
	repo := r.Context().Value("repo").(*models.Repo)
	num, _ := strconv.Atoi(mux.Vars(r)["number"])
	pid, _ := strconv.Atoi(mux.Vars(r)["pid"])

	logrus.Debugf("GetProcLogs: repo(%d) num(%d) pid(%d)", repo.ID, num, pid)
	store := s.Manager.Store()
	build, err := store.GetBuildNumber(repo, num)
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, nil, fmt.Sprintf("build not found: %s", err))
		return
	}

	proc, err := store.ProcFind(build, pid)
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, nil, fmt.Sprintf("proc not found: %s", err))
		return
	}

	rc, err := store.LogFind(proc)
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_NOT_FOUND, nil, fmt.Sprintf("log not found: %s", err))
		return
	}

	defer rc.Close()

	w.Header().Set("Content-Type", "application/json")
	io.Copy(w, rc)
}
