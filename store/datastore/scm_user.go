package datastore

import (
	"github.com/BoxLinker/cicd/models"
	"github.com/BoxLinker/cicd/store/datastore/sql"
	"github.com/russross/meddler"
	"github.com/Sirupsen/logrus"
	"time"
)

func (db *datastore) SaveSCMUser(user *models.SCMUser) error {
	logrus.Debugf("SaveSCMUser (%+v)", user)
	user.Created = time.Now()
	user.CreatedUnix = user.Created.Unix()
	user.Updated = time.Now()
	user.UpdatedUnix = user.Updated.Unix()
	return meddler.Insert(db, TableSCMUsers, user)
}

func (db *datastore) GetSCMUserByUCenterID(uCenterID string, scm string) *models.SCMUser {
	stmt := sql.Lookup(db.driver, SQLSCMUsersFindByUCenterID)
	u := new(models.SCMUser)
	if err := meddler.QueryRow(db, u, stmt, uCenterID, scm); err != nil {
		logrus.Errorf("GetSCMUserByUCenterID err (%s)", err.Error())
		return nil
	}
	u.Created = time.Unix(u.CreatedUnix, 0)
	u.Updated = time.Unix(u.UpdatedUnix, 0)
	logrus.Debugf("GetSCMUserByUCenterID (%+v)", u)
	return u
}

func (db *datastore) GetSCMUserByID(id int64) {}

func (db *datastore) UpdateSCMUser(user *models.SCMUser) error {
	user.Updated = time.Now()
	user.UpdatedUnix = user.Updated.Unix()
	return meddler.Update(db, TableSCMUsers, user)
}