package manager

import (
	"github.com/BoxLinker/cicd/models"
)

func (m *DefaultManager) SaveSCMUser(user *models.SCMUser) error {
	return m.dataStore.SaveSCMUser(user)
}

func (m *DefaultManager) GetSCMUserByUCenterID(uCenterID int64, scm models.SCMType) *models.SCMUser {
	return m.dataStore.GetSCMUserByUCenterID(uCenterID, scm)
}

func (m *DefaultManager) UpdateSCMUser(user *models.SCMUser) error {
	return m.dataStore.UpdateSCMUser(user)
}