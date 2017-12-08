package manager

import (
	"github.com/BoxLinker/cicd/models"
)

func (m *DefaultManager) SaveUser(user *models.User) error {
	return m.dataStore.SaveUser(user)
}

func (m *DefaultManager) GetUserByUCenterID(uCenterID string, scm string) *models.User {
	return m.dataStore.GetUserByUCenterID(uCenterID, scm)
}

func (m *DefaultManager) GetUserByIDAndSCM(id int64, scm string) (*models.User, error) {
	return m.dataStore.GetUserByIDAndSCM(id, scm)
}

func (m *DefaultManager) UpdateUser(user *models.User) error {
	return m.dataStore.UpdateUser(user)
}