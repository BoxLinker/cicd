package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/cabernety/gopkg/httplib"
	"github.com/gorilla/mux"

	"github.com/BoxLinker/cicd/models"
)

// GetBuild 根据 repo 和 build_number 获取 build 信息
func (s *Server) GetBuild(w http.ResponseWriter, r *http.Request) {
	buildNum, _ := strconv.Atoi(mux.Vars(r)["number"])
	repo := r.Context().Value("repo").(*models.Repo)
	build, err := s.Manager.Store().GetBuildNumber(repo, buildNum)
	if err != nil {
		httplib.Resp(w, httplib.STATUS_NOT_FOUND, nil, fmt.Sprintf("build not found: %v", err))
		return
	}
	httplib.Resp(w, httplib.STATUS_OK, build)
}
