package server

import (
	"net/http"

	"github.com/cabernety/gopkg/httplib"
)

func (s *Server) GetScms(w http.ResponseWriter, r *http.Request) {
	uCenterID := s.getCtxUserID(r)
	results := s.Manager.Store().GetUserScms(uCenterID)
	httplib.Resp(w, httplib.STATUS_OK, results)
}
