package datastore

import (
	"github.com/BoxLinker/cicd/models"
	"github.com/BoxLinker/cicd/store/datastore/sql"
	"github.com/russross/meddler"
	"github.com/Sirupsen/logrus"
	"time"
)

func (db *datastore) RepoList(u *models.SCMUser) ([]*models.Repo) {
	stmt := sql.Lookup(db.driver, SQLQueryReposByUserID)
	data := make([]*models.Repo, 0)
	if err := meddler.QueryAll(db, &data, stmt, u.ID); err != nil {
		logrus.Errorf("DataStore RepoList err (%s)", err.Error())
		return nil
	}
	for _, repo := range data {
		repo.Created = time.Unix(repo.CreatedUnix, 0)
		repo.Updated = time.Unix(repo.UpdatedUnix, 0)
	}
	return data
}

func (db *datastore) RepoBatch(user *models.SCMUser, repos []*models.Repo) error {
	stmt := sql.Lookup(db.driver, SQLRepoBatch)
	for _, repo := range repos {
		repo.UserID = user.ID
		repo.Created = time.Now()
		repo.CreatedUnix = repo.Created.Unix()
		repo.Updated = time.Now()
		repo.UpdatedUnix = repo.Updated.Unix()
		if _, err := db.Exec(stmt,
			repo.UserID,
			repo.Owner,
			repo.Name,
			repo.FullName,
			repo.SCM,
			repo.Link,
			repo.Clone,
			repo.Branch,
			repo.IsPrivate,
			repo.CreatedUnix,
			repo.UpdatedUnix,
		); err != nil {
			return err
		}
	}
	return nil
}