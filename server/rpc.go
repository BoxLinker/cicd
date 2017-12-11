package server

import (
	"github.com/BoxLinker/cicd/scm"
	"github.com/BoxLinker/cicd/queue"
	"github.com/BoxLinker/cicd/logging"
	"github.com/BoxLinker/cicd/pubsub"
	"github.com/BoxLinker/cicd/store"

	oldcontext "golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
	"github.com/Sirupsen/logrus"
	"github.com/BoxLinker/cicd/pipeline/rpc"
	"github.com/BoxLinker/cicd/modules/expr"
	"context"
	"encoding/json"
	"strconv"
	"github.com/BoxLinker/cicd/models"
	"bytes"
	"fmt"
	"github.com/BoxLinker/cicd/pipeline/rpc/proto"
)

type BoxCIServer struct {
	SCM scm.SCM
	Queue queue.Queue
	Logger logging.Log
	Pubsub pubsub.Publisher
	Store store.Store
	Host string
}

func (s *BoxCIServer) Next(c oldcontext.Context, req *proto.NextRequest) (*proto.NextReply, error) {
	peer := RPC{
		scm: s.SCM,
		store:  s.Store,
		queue:  s.Queue,
		pubsub: s.Pubsub,
		logger: s.Logger,
		host:   s.Host,
	}
	filter := rpc.Filter{
		Labels: req.GetFilter().GetLabels(),
	}

	res := new(proto.NextReply)
	pipeline, err := peer.Next(c, filter)
	if err != nil {
		return res, err
	}
	if pipeline == nil {
		return res, err
	}

	res.Pipeline = new(proto.Pipeline)
	res.Pipeline.Id = pipeline.ID
	res.Pipeline.Timeout = pipeline.Timeout
	res.Pipeline.Payload, _ = json.Marshal(pipeline.Config)

	return res, err
}

func (s *BoxCIServer) Init(c oldcontext.Context, req *proto.InitRequest) (*proto.Empty, error) {
	peer := RPC{
		scm: s.SCM,
		store:  s.Store,
		queue:  s.Queue,
		pubsub: s.Pubsub,
		logger: s.Logger,
		host:   s.Host,
	}
	state := rpc.State{
		Error:    req.GetState().GetError(),
		ExitCode: int(req.GetState().GetExitCode()),
		Finished: req.GetState().GetFinished(),
		Started:  req.GetState().GetStarted(),
		Proc:     req.GetState().GetName(),
		Exited:   req.GetState().GetExited(),
	}
	res := new(proto.Empty)
	err := peer.Init(c, req.GetId(), state)
	return res, err
}

func (s *BoxCIServer) Update(c oldcontext.Context, req *proto.UpdateRequest) (*proto.Empty, error) {
	peer := RPC{
		scm: s.SCM,
		store:  s.Store,
		queue:  s.Queue,
		pubsub: s.Pubsub,
		logger: s.Logger,
		host:   s.Host,
	}
	state := rpc.State{
		Error:    req.GetState().GetError(),
		ExitCode: int(req.GetState().GetExitCode()),
		Finished: req.GetState().GetFinished(),
		Started:  req.GetState().GetStarted(),
		Proc:     req.GetState().GetName(),
		Exited:   req.GetState().GetExited(),
	}
	res := new(proto.Empty)
	err := peer.Update(c, req.GetId(), state)
	return res, err
}

func (s *BoxCIServer) Upload(c oldcontext.Context, req *proto.UploadRequest) (*proto.Empty, error) {
	peer := RPC{
		scm: s.SCM,
		store:  s.Store,
		queue:  s.Queue,
		pubsub: s.Pubsub,
		logger: s.Logger,
		host:   s.Host,
	}
	file := &rpc.File{
		Data: req.GetFile().GetData(),
		Mime: req.GetFile().GetMime(),
		Name: req.GetFile().GetName(),
		Proc: req.GetFile().GetProc(),
		Size: int(req.GetFile().GetSize()),
		Time: req.GetFile().GetTime(),
		Meta: req.GetFile().GetMeta(),
	}

	res := new(proto.Empty)
	err := peer.Upload(c, req.GetId(), file)
	return res, err
}


func (s *BoxCIServer) Done(c oldcontext.Context, req *proto.DoneRequest) (*proto.Empty, error) {
	peer := RPC{
		scm: s.SCM,
		store:  s.Store,
		queue:  s.Queue,
		pubsub: s.Pubsub,
		logger: s.Logger,
		host:   s.Host,
	}
	state := rpc.State{
		Error:    req.GetState().GetError(),
		ExitCode: int(req.GetState().GetExitCode()),
		Finished: req.GetState().GetFinished(),
		Started:  req.GetState().GetStarted(),
		Proc:     req.GetState().GetName(),
		Exited:   req.GetState().GetExited(),
	}
	res := new(proto.Empty)
	err := peer.Done(c, req.GetId(), state)
	return res, err
}

func (s *BoxCIServer) Wait(c oldcontext.Context, req *proto.WaitRequest) (*proto.Empty, error) {
	peer := RPC{
		scm: s.SCM,
		store:  s.Store,
		queue:  s.Queue,
		pubsub: s.Pubsub,
		logger: s.Logger,
		host:   s.Host,
	}
	res := new(proto.Empty)
	err := peer.Wait(c, req.GetId())
	return res, err
}

func (s *BoxCIServer) Extend(c oldcontext.Context, req *proto.ExtendRequest) (*proto.Empty, error) {
	peer := RPC{
		scm: s.SCM,
		store:  s.Store,
		queue:  s.Queue,
		pubsub: s.Pubsub,
		logger: s.Logger,
		host:   s.Host,
	}
	res := new(proto.Empty)
	err := peer.Extend(c, req.GetId())
	return res, err
}

func (s *BoxCIServer) Log(c oldcontext.Context, req *proto.LogRequest) (*proto.Empty, error) {
	peer := RPC{
		scm: s.SCM,
		store:  s.Store,
		queue:  s.Queue,
		pubsub: s.Pubsub,
		logger: s.Logger,
		host:   s.Host,
	}
	line := &rpc.Line{
		Out:  req.GetLine().GetOut(),
		Pos:  int(req.GetLine().GetPos()),
		Time: req.GetLine().GetTime(),
		Proc: req.GetLine().GetProc(),
	}
	res := new(proto.Empty)
	err := peer.Log(c, req.GetId(), line)
	return res, err
}



func createFilterFunc(filter rpc.Filter) (queue.Filter, error) {
	var st *expr.Selector
	var err error

	if filter.Expr != "" {
		st, err = expr.ParseString(filter.Expr)
		if err != nil {
			return nil, err
		}
	}

	return func(task *queue.Task) bool {
		if st != nil {
			match, _ := st.Eval(expr.NewRow(task.Labels))
			return match
		}

		for k, v := range filter.Labels {
			if task.Labels[k] != v {
				return false
			}
		}
		return true
	}, nil
}

type RPC struct {
	scm scm.SCM
	queue queue.Queue
	pubsub pubsub.Publisher
	logger logging.Log
	store store.Store
	host string
}


func (s *RPC) Next(c context.Context, filter rpc.Filter) (*rpc.Pipeline, error) {
	metadata, ok := metadata.FromIncomingContext(c)
	if ok {
		hostname, ok := metadata["hostname"]
		if ok && len(hostname) != 0 {
			logrus.Debugf("agent connected: %s: polling", hostname[0])
		}
	}

	fn, err := createFilterFunc(filter)
	if err != nil {
		return nil, err
	}
	task, err := s.queue.Poll(c, fn)
	if err != nil {
		return nil, err
	} else if task == nil {
		return nil, nil
	}
	pipeline := new(rpc.Pipeline)

	err = json.Unmarshal(task.Data, pipeline)
	return pipeline, err
}

func (s *RPC) Wait(c context.Context, id string) error {
	return s.queue.Wait(c, id)
}

func (s *RPC) Extend(c context.Context, id string) error {
	return s.queue.Extend(c, id)
}

func (s *RPC) Update(c context.Context, id string, state rpc.State) error {
	procID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	pproc, err := s.store.ProcLoad(procID)
	if err != nil {
		logrus.Errorf("error: rpc.update: cannot find pproc with id %d: %s", procID, err)
		return err
	}

	build, err := s.store.GetBuild(pproc.BuildID)
	if err != nil {
		logrus.Errorf("error: cannot find build with id %d: %s", pproc.BuildID, err)
		return err
	}

	proc, err := s.store.ProcChild(build, pproc.PID, state.Proc)
	if err != nil {
		logrus.Errorf("error: cannot find proc with name %s: %s", state.Proc, err)
		return err
	}

	metadata, ok := metadata.FromIncomingContext(c)
	if ok {
		hostname, ok := metadata["hostname"]
		if ok && len(hostname) != 0 {
			proc.Machine = hostname[0]
		}
	}

	repo, err := s.store.GetRepo(build.RepoID)
	if err != nil {
		logrus.Errorf("error: cannot find repo with id %d: %s", build.RepoID, err)
		return err
	}

	if state.Exited {
		proc.Stopped = state.Finished
		proc.ExitCode = state.ExitCode
		proc.Error = state.Error
		proc.State = models.StatusSuccess
		if state.ExitCode != 0 || state.Error != "" {
			proc.State = models.StatusFailure
		}

		if state.ExitCode == 137 {
			proc.State = models.StatusKilled
		}
	} else {
		proc.Started = state.Started
		proc.State = models.StatusRunning
	}

	if err := s.store.ProcUpdate(proc); err != nil {
		logrus.Errorf("error: rpc.update: cannot update proc: %s", err)
	}

	build.Procs, _ = s.store.ProcList(build)
	build.Procs = models.Tree(build.Procs)
	message := pubsub.Message{
		Labels: map[string]string{
			"repo": repo.FullName,
			"private": strconv.FormatBool(repo.IsPrivate),
		},
	}
	message.Data, _ = json.Marshal(models.Event{
		Repo: *repo,
		Build: *build,
	})
	s.pubsub.Publish(c, "topic/events", message)

	return nil
}

func (s *RPC) Upload(c context.Context, id string, file *rpc.File) error {
	procID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	pproc, err := s.store.ProcLoad(procID)
	if err != nil {
		logrus.Errorf("error: cannot find parent proc with id %d: %s", procID, err)
		return err
	}

	build, err := s.store.GetBuild(pproc.BuildID)
	if err != nil {
		logrus.Errorf("error: cannot find build with id %d: %s", pproc.BuildID, err)
		return err
	}

	proc, err := s.store.ProcChild(build, pproc.PID, file.Proc)
	if err != nil {
		logrus.Errorf("error: cannot find child proc with name: %s: %s", file.Proc, err)
		return err
	}

	if file.Mime == "application/json+logs" {
		return s.store.LogSave(proc, bytes.NewBuffer(file.Data))
	}

	report := &models.File{
		BuildID: proc.BuildID,
		ProcID:  proc.ID,
		PID:     proc.PID,
		Mime:    file.Mime,
		Name:    file.Name,
		Size:    file.Size,
		Time:    file.Time,
	}
	if d, ok := file.Meta["X-Tests-Passed"]; ok {
		report.Passed, _ = strconv.Atoi(d)
	}
	if d, ok := file.Meta["X-Tests-Failed"]; ok {
		report.Failed, _ = strconv.Atoi(d)
	}
	if d, ok := file.Meta["X-Tests-Skipped"]; ok {
		report.Skipped, _ = strconv.Atoi(d)
	}

	if d, ok := file.Meta["X-Checks-Passed"]; ok {
		report.Passed, _ = strconv.Atoi(d)
	}
	if d, ok := file.Meta["X-Checks-Failed"]; ok {
		report.Failed, _ = strconv.Atoi(d)
	}

	if d, ok := file.Meta["X-Coverage-Lines"]; ok {
		report.Passed, _ = strconv.Atoi(d)
	}
	if d, ok := file.Meta["X-Coverage-Total"]; ok {
		if total, _ := strconv.Atoi(d); total != 0 {
			report.Failed = total - report.Passed
		}
	}

	return s.store.FileCreate(report, bytes.NewBuffer(file.Data))
}


// Init implements the rpc.Init function
func (s *RPC) Init(c context.Context, id string, state rpc.State) error {
	procID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	proc, err := s.store.ProcLoad(procID)
	if err != nil {
		logrus.Errorf("error: cannot find proc with id %d: %s", procID, err)
		return err
	}
	metadata, ok := metadata.FromIncomingContext(c)
	if ok {
		hostname, ok := metadata["hostname"]
		if ok && len(hostname) != 0 {
			proc.Machine = hostname[0]
		}
	}

	build, err := s.store.GetBuild(proc.BuildID)
	if err != nil {
		logrus.Errorf("error: cannot find build with id %d: %s", proc.BuildID, err)
		return err
	}

	repo, err := s.store.GetRepo(build.RepoID)
	if err != nil {
		logrus.Errorf("error: cannot find repo with id %d: %s", build.RepoID, err)
		return err
	}

	if build.Status == models.StatusPending {
		build.Status = models.StatusRunning
		build.Started = state.Started
		if err := s.store.UpdateBuild(build); err != nil {
			logrus.Errorf("error: init: cannot update build_id %d state: %s", build.ID, err)
		}
	}

	defer func() {
		build.Procs, _ = s.store.ProcList(build)
		message := pubsub.Message{
			Labels: map[string]string{
				"repo":    repo.FullName,
				"private": strconv.FormatBool(repo.IsPrivate),
			},
		}
		message.Data, _ = json.Marshal(models.Event{
			Repo:  *repo,
			Build: *build,
		})
		s.pubsub.Publish(c, "topic/events", message)
	}()

	proc.Started = state.Started
	proc.State = models.StatusRunning
	return s.store.ProcUpdate(proc)
}


// Done implements the rpc.Done function
func (s *RPC) Done(c context.Context, id string, state rpc.State) error {
	procID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	proc, err := s.store.ProcLoad(procID)
	if err != nil {
		logrus.Errorf("error: cannot find proc with id %d: %s", procID, err)
		return err
	}

	build, err := s.store.GetBuild(proc.BuildID)
	if err != nil {
		logrus.Errorf("error: cannot find build with id %d: %s", proc.BuildID, err)
		return err
	}

	repo, err := s.store.GetRepo(build.RepoID)
	if err != nil {
		logrus.Errorf("error: cannot find repo with id %d: %s", build.RepoID, err)
		return err
	}

	proc.Stopped = state.Finished
	proc.Error = state.Error
	proc.ExitCode = state.ExitCode
	proc.State = models.StatusSuccess
	if proc.ExitCode != 0 || proc.Error != "" {
		proc.State = models.StatusFailure
	}
	if err := s.store.ProcUpdate(proc); err != nil {
		logrus.Errorf("error: done: cannot update proc_id %d state: %s", procID, err)
	}

	if err := s.queue.Done(c, id); err != nil {
		logrus.Errorf("error: done: cannot ack proc_id %d: %s", procID, err)
	}

	// TODO handle this error
	procs, _ := s.store.ProcList(build)
	for _, p := range procs {
		if p.Running() && p.PPID == proc.PID {
			p.State = models.StatusSkipped
			if p.Started != 0 {
				p.State = models.StatusSuccess // for deamons that are killed
				p.Stopped = proc.Stopped
			}
			if err := s.store.ProcUpdate(p); err != nil {
				logrus.Errorf("error: done: cannot update proc_id %d child state: %s", p.ID, err)
			}
		}
	}

	running := false
	status := models.StatusSuccess
	for _, p := range procs {
		if p.PPID == 0 {
			if p.Running() {
				running = true
			}
			if p.Failing() {
				status = p.State
			}
		}
	}
	if !running {
		build.Status = status
		build.Finished = proc.Stopped
		if err := s.store.UpdateBuild(build); err != nil {
			logrus.Errorf("error: done: cannot update build_id %d final state: %s", build.ID, err)
		}

		// update the status
		user, err := s.store.GetUserByIDAndSCM(repo.UserID, repo.SCM)
		if err == nil {
			if refresher, ok := s.scm.(scm.Refresher); ok {
				ok, _ := refresher.Refresh(user)
				if ok {
					s.store.UpdateUser(user)
				}
			}
			uri := fmt.Sprintf("%s/%s/%d", s.host, repo.FullName, build.Number)
			err = s.scm.Status(user, repo, build, uri)
			if err != nil {
				logrus.Errorf("error setting commit status for %s/%d", repo.FullName, build.Number)
			}
		}
	}

	if err := s.logger.Close(c, id); err != nil {
		logrus.Errorf("error: done: cannot close build_id %d logger: %s", proc.ID, err)
	}

	build.Procs = models.Tree(procs)
	message := pubsub.Message{
		Labels: map[string]string{
			"repo":    repo.FullName,
			"private": strconv.FormatBool(repo.IsPrivate),
		},
	}
	message.Data, _ = json.Marshal(models.Event{
		Repo:  *repo,
		Build: *build,
	})
	s.pubsub.Publish(c, "topic/events", message)

	return nil
}

// Log implements the rpc.Log function
func (s *RPC) Log(c context.Context, id string, line *rpc.Line) error {
	entry := new(logging.Entry)
	entry.Data, _ = json.Marshal(line)
	s.logger.Write(c, id, entry)
	return nil
}
