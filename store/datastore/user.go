package datastore

import (
	"time"

	"github.com/BoxLinker/cicd/models"
	"github.com/BoxLinker/cicd/store/datastore/sql"
	"github.com/Sirupsen/logrus"
	"github.com/russross/meddler"
)

func (db *datastore) SaveUser(user *models.User) error {
	logrus.Debugf("SaveUser (%+v)", user)
	user.Created = time.Now()
	user.CreatedUnix = user.Created.Unix()
	user.Updated = time.Now()
	user.UpdatedUnix = user.Updated.Unix()
	if err := meddler.Insert(db, TableSCMUsers, user); err != nil {
		logrus.Errorf("SaveUser error: %s", err)
		return err
	}
	return nil
}

func (db *datastore) GetUserByUCenterID(uCenterID string, scm string) *models.User {
	stmt := sql.Lookup(db.driver, SQLSCMUsersFindByUCenterID)
	u := new(models.User)
	if err := meddler.QueryRow(db, u, stmt, uCenterID, scm); err != nil {
		logrus.Errorf("GetUserByUCenterID(%s) err (%s)", uCenterID, err.Error())
		return nil
	}
	u.Created = time.Unix(u.CreatedUnix, 0)
	u.Updated = time.Unix(u.UpdatedUnix, 0)
	logrus.Debugf("GetUserByUCenterID (%+v)", u)
	return u
}

func (db *datastore) GetUserByIDAndSCM(id int64, scm string) (*models.User, error) {
	stmt := sql.Lookup(db.driver, SQLUserFindByIDSCM)
	u := new(models.User)
	if err := meddler.QueryRow(db, u, stmt, id, scm); err != nil {
		logrus.Errorf("GetUserByIDAndSCM err (%s)", err.Error())
		return nil, err
	}
	return u, nil
}

func (db *datastore) UpdateUser(user *models.User) error {
	user.Updated = time.Now()
	user.UpdatedUnix = user.Updated.Unix()
	return meddler.Update(db, TableSCMUsers, user)
}
