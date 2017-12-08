package rpc

import "github.com/BoxLinker/cicd/pipeline/backend"

type (
	Pipeline struct {
		ID string `json:"id"`
		Config *backend.Config `json:"config"`
		Timeout int64 `json:"timeout"`
	}
)