package manager

import (
	"github.com/BoxLinker/cicd/models"
	"github.com/cabernety/gopkg/httplib"
)

func (m *DefaultManager) RepoBatch(user *models.User, repos []*models.Repo) error {
	return m.dataStore.RepoBatch(user, repos)
}

// TODO 分页查询
func (m *DefaultManager) QueryRepos(u *models.User, pc *httplib.PageConfig) []*models.Repo {
	return m.dataStore.RepoList(u)
}

func (m *DefaultManager) GetRepoOwnerName(owner, repoName, scm string) (*models.Repo, error) {
	return m.dataStore.GetRepoOwnerName(owner, repoName, scm)
}
