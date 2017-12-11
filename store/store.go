package store

import (
	"github.com/BoxLinker/cicd/models"
	"io"
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

	GetBuild(int64) (*models.Build, error)
	CreateBuild(*models.Build, ...*models.Proc) error
	UpdateBuild(*models.Build) error

	ProcCreate([]*models.Proc) error
	ProcLoad(int64) (*models.Proc, error)
	ProcChild(build *models.Build, pid int, child string) (*models.Proc, error)
	ProcUpdate(proc *models.Proc) error
	ProcList(build *models.Build) ([]*models.Proc, error)

	LogSave(proc *models.Proc, reader io.Reader) error

	FileCreate(file *models.File, reader io.Reader) error
	FileList(build *models.Build) ([]*models.File, error)
	FileFind(proc *models.Proc, name string) (*models.File, error)
	FileRead(proc *models.Proc, name string) (io.ReadCloser, error)
}
