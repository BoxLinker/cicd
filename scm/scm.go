package scm

import (
	"net/http"
	"github.com/BoxLinker/cicd/models"
)

type SCM interface{
	Authorize(w http.ResponseWriter, r *http.Request, stateParam string) (*models.User, error)

	Repos(u *models.User) ([]*models.Repo, error)

	File(u *models.User, r *models.Repo, b *models.Build, f string) ([]byte, error)
	FileRef(u *models.User, r *models.Repo, ref, f string) ([]byte, error)

	// Status sends the commit status to the remote system.
	Status(u *models.User, r *models.Repo, b *models.Build, link string) error

	Hook(r *http.Request) (*models.Repo, *models.Build, error)
}
