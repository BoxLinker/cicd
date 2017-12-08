package store

import (
	"github.com/BoxLinker/cicd/models"
)

type Store interface {
	SaveUser(user *models.User) error
	GetUserByUCenterID(uCenterID string, scm string) *models.User
	UpdateUser(user *models.User) error
	GetUserByIDAndSCM(id int64, scm string) (*models.User, error)

	GetRepo(id int64) (*models.Repo, error)
	RepoList(u *models.User) ([]*models.Repo)
	RepoBatch(user *models.User, repos []*models.Repo) error
	GetRepoOwnerName(owner, repoName string) (*models.Repo, error)

	ConfigLoad(int64) (*models.Config, error)
	ConfigFind(*models.Repo, string) (*models.Config, error)
	ConfigFindApproved(*models.Config) (bool, error)
	ConfigCreate(*models.Config) error

	TaskList() ([]*models.Task, error)
	TaskInsert(*models.Task) error
	TaskDelete(string) error

	CreateBuild(*models.Build, ...*models.Proc) error
	UpdateBuild(*models.Build) error

	ProcCreate([]*models.Proc) error
}
