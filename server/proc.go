package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/cabernety/gopkg/httplib"
	"github.com/gorilla/mux"
)

func (s *Server) GetProcs(w http.ResponseWriter, r *http.Request) {
	buildID, _ := strconv.Atoi(mux.Vars(r)["build_id"])
	build, err := s.Manager.Store().GetBuild(int64(buildID))
	if err != nil {
		httplib.Resp(w, httplib.STATUS_NOT_FOUND, nil, fmt.Sprintf("build not found: %v", err))
		return
	}
	procs, err := s.Manager.Store().ProcList(build)
	if err != nil {
		httplib.Resp(w, httplib.STATUS_NOT_FOUND, nil, fmt.Sprintf("procs not found: %v", err))
		return
	}
	httplib.Resp(w, httplib.STATUS_OK, procs)
}
