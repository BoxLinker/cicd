package manager

import (
	"github.com/BoxLinker/cicd/models"
	"github.com/BoxLinker/boxlinker-api"
)

func (m *DefaultManager) RepoBatch(repos []*models.Repo) error {
	return m.dataStore.RepoBatch(repos)
}

// TODO 分页查询
func (m *DefaultManager) QueryRepos(u *models.SCMUser, pc *boxlinker.PageConfig) ([]*models.Repo) {
	return m.dataStore.RepoList(u)
}