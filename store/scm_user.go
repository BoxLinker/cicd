package store

import (
	"github.com/BoxLinker/cicd/models"
	"github.com/BoxLinker/cicd/store/sql"
	"github.com/russross/meddler"
	"github.com/Sirupsen/logrus"
)

func (db *DataStore) SaveSCMUser(user *models.SCMUser) error {
	return meddler.Insert(db, TableSCMUsers, user)
}

func (db *DataStore) GetSCMUserByUCenterID(uCenterID int64, scm models.SCMType) *models.SCMUser {
	stmt := sql.Lookup(db.driver, SQLSCMUsersFindByUCenterID)
	u := new(models.SCMUser)
	if err := meddler.QueryRow(db, u, stmt, uCenterID, scm); err != nil {
		logrus.Errorf("GetSCMUserByUCenterID err (%s)", err.Error())
		return nil
	}
	return u
}

func (db *DataStore) UpdateSCMUser(user *models.SCMUser) error {
	return meddler.Update(db, TableSCMUsers, user)
}