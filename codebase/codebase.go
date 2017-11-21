package codebase

import (
	"net/http"
	"github.com/BoxLinker/cicd/models"
)

type CodeBase interface{
	Authorize(w http.ResponseWriter, r *http.Request) (*models.User, error)
}
