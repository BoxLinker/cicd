package manager

import (
	"github.com/BoxLinker/cicd/models"
	"github.com/BoxLinker/boxlinker-api"
	"github.com/Sirupsen/logrus"
)

func (m *DefaultManager) SaveRepos(repos []*models.Repo) error {
	sess := m.DBEngine.NewSession()
	defer sess.Close()
	if _, err := sess.InsertMulti(&repos); err != nil {
		sess.Rollback()
		return err
	}
	return sess.Commit()
}

func (m *DefaultManager) QueryRepos(u *models.CodeBaseUser, pc *boxlinker.PageConfig) (repos []*models.Repo) {
	if u == nil || u.UserID == "" {
		return nil
	}
	if err := m.DBEngine.Where("user_id = ?", u.ID).Limit(pc.Limit(), pc.Offset()).Find(&repos); err != nil {
		logrus.Errorf("QueryRepos err (%s)", err.Error())
		return nil
	}
	return
}