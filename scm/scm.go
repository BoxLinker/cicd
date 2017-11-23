package scm

import (
	"net/http"
	"github.com/BoxLinker/cicd/models"
)

type SCM interface{
	Authorize(w http.ResponseWriter, r *http.Request, stateParam string) (*models.SCMUser, error)

	Repos(u *models.SCMUser) ([]*models.Repo, error)
}
