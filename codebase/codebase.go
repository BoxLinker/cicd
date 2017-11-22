package codebase

import (
	"net/http"
	"github.com/BoxLinker/cicd/models"
)

type CodeBase interface{
	Authorize(w http.ResponseWriter, r *http.Request, stateParam string) (*models.CodeBaseUser, error)

	Repos(u *models.CodeBaseUser) ([]*models.Repo, error)
}
