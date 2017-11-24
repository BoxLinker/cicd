package server

import (
	"net/http"
	"github.com/BoxLinker/boxlinker-api"
	"strconv"
	"github.com/Sirupsen/logrus"
)

func (s *Server) GetRepos(w http.ResponseWriter, r *http.Request) {
	flush, _ := strconv.ParseBool(boxlinker.GetQueryParam(r, "flush"))
	pc := boxlinker.ParsePageConfig(r)
	u := s.getUserInfo(r)
	logrus.Debugf("GetRepos (%s)", u.SCM)

	if flush {
		if repos, err := s.Manager.GetSCM(u.SCM).Repos(u); err != nil {
			boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
			return
		} else {
			if err := s.Manager.RepoBatch(u, repos); err != nil {
				boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
				return
			}
		}
	}

	boxlinker.Resp(w, boxlinker.STATUS_OK, pc.PaginationResult(s.Manager.QueryRepos(u, &pc)))
}
