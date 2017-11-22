package manager

import (
	"github.com/BoxLinker/cicd/models"
	"github.com/Sirupsen/logrus"
)

func (m *DefaultManager) SaveCodeBaseUser(user *models.CodeBaseUser) error {
	sess := m.DBEngine.NewSession()
	defer sess.Close()
	_, err := sess.Insert(user)
	if err != nil {
		sess.Rollback()
		return err
	}
	return sess.Commit()
}

func (m *DefaultManager) GetCodeBaseUserByUserID(userID string) *models.CodeBaseUser {
	u := &models.CodeBaseUser{
		UserID: userID,
	}
	if has, err := m.DBEngine.Get(u); !has {
		logrus.Errorf("GetCodeBaseUserByUserID err (%s)", err.Error())
		return nil
	}
	return u
}

func (m *DefaultManager) IsCodeBaseUserExists(userID, kind string) (bool, error) {
	u := &models.CodeBaseUser{
		UserID: userID,
		Kind: kind,
	}
	return m.DBEngine.Get(u)
}

func (m *DefaultManager) UpdateCodeBaseUser(user *models.CodeBaseUser) error {
	sess := m.DBEngine.NewSession()
	defer sess.Close()
	u := &models.CodeBaseUser{
		Login: user.Login,
		Email: user.Email,
		AccessToken: user.AccessToken,
	}
	_, err := sess.Update(u)
	if err != nil {
		sess.Rollback()
		return err
	}
	return nil
}