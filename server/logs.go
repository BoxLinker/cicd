package server

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/BoxLinker/boxlinker-api"
	"github.com/BoxLinker/cicd/logging"
	"github.com/BoxLinker/cicd/models"
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

func (s *Server) LogStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	cw, ok := w.(http.CloseNotifier)
	if !ok {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, "Streaming not supported CloseNotifier")
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, "Streaming not supported Flusher")
		return
	}

	io.WriteString(w, ": ping\n\n")
	flusher.Flush()

	repo := r.Context().Value("repo").(*models.Repo)
	buildn, _ := strconv.Atoi(mux.Vars(r)["build"])
	jobn, _ := strconv.Atoi(mux.Vars(r)["number"])
	store := s.Manager.Store()
	build, err := store.GetBuildNumber(repo, buildn)
	if err != nil {
		logrus.Errorf("stream can not get build number: %v", err)
		io.WriteString(w, "event error:\ndata: build not found\n\n")
		return
	}

	proc, err := store.ProcFind(build, jobn)
	if err != nil {
		logrus.Errorf("stream can not get proc number: %v", err)
		io.WriteString(w, "event error:\ndata: process not found\n\n")
		return
	}

	if proc.State != models.StatusRunning {
		logrus.Debugln("stream not found")
		io.WriteString(w, "event error:\ndata: stream not found\n\n")
		return
	}

	logc := make(chan []byte, 10)
	ctx, cancel := context.WithCancel(context.Background())

	logrus.Debugln("log stream: connection opened")

	defer func() {
		cancel()
		close(logc)
		logrus.Debugln("log stream: connection closed")
	}()

	go func() {
		s.Manager.Logs().Tail(ctx, fmt.Sprint(proc.ID), func(entries ...*logging.Entry) {
			for _, entry := range entries {
				select {
				case <-ctx.Done():
					return
				default:
					logc <- entry.Data
				}
			}
		})

		io.WriteString(w, "event: error\ndata: eof\n\n")
		cancel()
	}()

	id := 1
	last, _ := strconv.Atoi(r.Header.Get("Last-Event-ID"))
	if last != 0 {
		logrus.Debugf("log stream reconnect: last-event-id: %d", last)
	}

	for {
		select {
		case <-time.After(time.Hour):
			return
		case <-cw.CloseNotify():
			return
		case <-time.After(time.Second * 30):
			io.WriteString(w, ": ping\n\n")
			flusher.Flush()
		case buf, ok := <-logc:
			if ok {
				if id > last {
					io.WriteString(w, "id: "+strconv.Itoa(id))
					io.WriteString(w, "\n")
					io.WriteString(w, "data: ")
					w.Write(buf)
					io.WriteString(w, "\n\n")
					flusher.Flush()
				}
				id++
			}
		}
	}

}

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
