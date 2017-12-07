package store

import (
	"github.com/BoxLinker/cicd/models"
)

type Store interface {
	SaveSCMUser(user *models.SCMUser) error
	GetSCMUserByUCenterID(uCenterID string, scm string) *models.SCMUser
	UpdateSCMUser(user *models.SCMUser) error

	RepoList(u *models.SCMUser) ([]*models.Repo)
	RepoBatch(user *models.SCMUser, repos []*models.Repo) error
	GetRepoOwnerName(owner, repoName string) (*models.Repo, error)

	TaskList() ([]*models.Task, error)
	TaskInsert(*models.Task) error
	TaskDelete(string) error
}
