package server

import (
	"github.com/BoxLinker/cicd/scm"
	"github.com/BoxLinker/cicd/queue"
	"github.com/BoxLinker/cicd/logging"
	"github.com/BoxLinker/cicd/pubsub"
	"github.com/BoxLinker/cicd/store"

	oldcontext "golang.org/x/net/context"
	"github.com/BoxLinker/cicd/pipeline/rpc/proto"
)

type RPCServer struct {
	SCM scm.SCM
	Queue queue.Queue
	Logger logging.Log
	Pubsub pubsub.Publisher
	Store store.Store
}

func (s *RPCServer) Next(c oldcontext.Context, req *proto)