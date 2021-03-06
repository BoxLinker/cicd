package store

import (
	"io"

	"github.com/BoxLinker/cicd/models"
	"github.com/cabernety/gopkg/httplib"
)

type RepoListOptions struct {
	Active     bool
	User       *models.User
	All        bool
	Pagination *httplib.PageConfig
}

type Store interface {
	SaveUser(user *models.User) error
	GetUserByUCenterID(uCenterID string, scm string) *models.User
	UpdateUser(user *models.User) error
	GetUserByIDAndSCM(id int64, scm string) (*models.User, error)
	GetUserScms(uCenterID string) []*models.User

	GetRepo(id int64) (*models.Repo, error)
	RepoList(opt *RepoListOptions) []*models.Repo
	UpdateRepo(repo *models.Repo) error
	RepoBatch(user *models.User, repos []*models.Repo) error
	RepoCount(user *models.User) int
	GetRepoOwnerName(owner, repoName, scm string) (*models.Repo, error)

	BranchBatch(repo *models.Repo, branches []*models.Branch) error
	BranchList(repo *models.Repo, limit, offset int) []*models.Branch

	ConfigLoad(int64) (*models.Config, error)
	ConfigFind(*models.Repo, string) (*models.Config, error)
	ConfigFindApproved(*models.Config) (bool, error)
	ConfigCreate(*models.Config) error

	TaskList() ([]*models.Task, error)
	TaskInsert(*models.Task) error
	TaskDelete(string) error

	GetBuild(int64) (*models.Build, error)
	GetBuildNumber(repo *models.Repo, num int) (*models.Build, error)
	BuildCount(repo *models.Repo) int
	QueryBuild(repo *models.Repo, pc *httplib.PageConfig) []*models.Build
	// gets the last build before build number N
	GetBuildLastBefore(repo *models.Repo, branch string, n int64) (*models.Build, error)
	CreateBuild(*models.Build, ...*models.Proc) error
	UpdateBuild(*models.Build) error
	SearchBuild(repo *models.Repo, search string, pc *httplib.PageConfig) []*models.Build
	QueryBranchBuild(repo *models.Repo, branch string) []*models.Build

	ProcCreate([]*models.Proc) error
	ProcFind(build *models.Build, pid int) (*models.Proc, error)
	ProcLoad(int64) (*models.Proc, error)
	ProcChild(build *models.Build, pid int, child string) (*models.Proc, error)
	ProcUpdate(proc *models.Proc) error
	ProcList(build *models.Build) ([]*models.Proc, error)

	LogSave(proc *models.Proc, reader io.Reader) error
	LogFind(proc *models.Proc) (io.ReadCloser, error)

	FileCreate(file *models.File, reader io.Reader) error
	FileList(build *models.Build) ([]*models.File, error)
	FileFind(proc *models.Proc, name string) (*models.File, error)
	FileRead(proc *models.Proc, name string) (io.ReadCloser, error)

	SecretFind(*models.Repo, string) (*models.Secret, error)
	SecretList(*models.Repo) ([]*models.Secret, error)
	SecretCreate(*models.Secret) error
	SecretUpdate(*models.Secret) error
	SecretDelete(*models.Secret) error

	RegistryFind(*models.Repo, string) (*models.Registry, error)
	RegistryList(*models.Repo) ([]*models.Registry, error)
	RegistryCreate(*models.Registry) error
	RegistryUpdate(*models.Registry) error
	RegistryDelete(*models.Registry) error
}
