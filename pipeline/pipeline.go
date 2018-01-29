package pipeline

import (
	"context"
	"time"

	"github.com/BoxLinker/cicd/pipeline/backend"
	"github.com/BoxLinker/cicd/pipeline/multipart"
	"github.com/Sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type (
	// pipeline 和 process 的状态
	State struct {
		// pipeline 的全局状态
		Pipeline struct {
			// pipeline 开始时间
			Time int64 `json:"time"`
			// pipeline 当前 step
			Step *backend.Step `json:"step"`
			// 当前 pipeline 错误
			Error error `json:"error"`
		}

		// 当前 process 状态
		Process *backend.State
	}
)

// Runtime is a configuration runtime.
type Runtime struct {
	err     error
	spec    *backend.Config
	engine  backend.Engine
	started int64

	ctx    context.Context
	tracer Tracer
	logger Logger
}

// New returns a new runtime using the specified runtime
// configuration and runtime engine.
func New(spec *backend.Config, opts ...Option) *Runtime {
	r := new(Runtime)
	r.spec = spec
	r.ctx = context.Background()
	for _, opts := range opts {
		opts(r)
	}
	return r
}

// Run starts the runtime and waits for it to complete.
func (r *Runtime) Run() error {
	defer func() {
		r.engine.Destroy(r.spec)
	}()

	r.started = time.Now().Unix()
	if err := r.engine.Setup(r.spec); err != nil {
		logrus.Errorf("engine setup err: %v", err)
		return err
	}

	for _, stage := range r.spec.Stages {
		select {
		case <-r.ctx.Done():
			return ErrCancel
		case err := <-r.execAll(stage.Steps):
			if err != nil {
				r.err = err
			}
		}
	}

	return r.err
}

//
//
//

func (r *Runtime) execAll(procs []*backend.Step) <-chan error {
	var g errgroup.Group
	done := make(chan error)

	for _, proc := range procs {
		proc := proc
		g.Go(func() error {
			return r.exec(proc)
		})
	}

	go func() {
		done <- g.Wait()
		close(done)
	}()
	return done
}

//
//
//

func (r *Runtime) exec(proc *backend.Step) error {
	switch {
	case r.err != nil && proc.OnFailure == false:
		return nil
	case r.err == nil && proc.OnSuccess == false:
		return nil
	}

	if r.tracer != nil {
		state := new(State)
		state.Pipeline.Time = r.started
		state.Pipeline.Error = r.err
		state.Pipeline.Step = proc
		state.Process = new(backend.State) // empty
		if err := r.tracer.Trace(state); err == ErrSkip {
			return nil
		} else if err != nil {
			return err
		}
	}

	if err := r.engine.Exec(proc); err != nil {
		logrus.Errorf("engine exec err: %v", err)
		return err
	}

	if r.logger != nil {
		rc, err := r.engine.Tail(proc)
		if err != nil {
			return err
		}

		go func() {
			r.logger.Log(proc, multipart.New(rc))
			rc.Close()
		}()
	}

	if proc.Detached {
		return nil
	}

	wait, err := r.engine.Wait(proc)
	if err != nil {
		return err
	}

	if r.tracer != nil {
		state := new(State)
		state.Pipeline.Time = r.started
		state.Pipeline.Error = r.err
		state.Pipeline.Step = proc
		state.Process = wait
		if err := r.tracer.Trace(state); err != nil {
			return err
		}
	}

	if wait.OOMKilled {
		return &OomError{
			Name: proc.Name,
			Code: wait.ExitCode,
		}
	} else if wait.ExitCode != 0 {
		return &ExitError{
			Name: proc.Name,
			Code: wait.ExitCode,
		}
	}
	return nil
}