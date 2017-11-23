package store

import (
	"github.com/BoxLinker/cicd/models"
	"github.com/BoxLinker/cicd/store/sql"
	"github.com/russross/meddler"
	"github.com/Sirupsen/logrus"
)

func (db *DataStore) RepoList(u *models.SCMUser) ([]*models.Repo) {
	stmt := sql.Lookup(db.driver, SQLQueryReposByUserID)
	data := make([]*models.Repo, 0)
	if err := meddler.QueryAll(db, &data, stmt, u.ID); err != nil {
		logrus.Errorf("DataStore RepoList err (%s)", err.Error())
		return nil
	}
	return data
}

func (db *DataStore) RepoBatch(repos []*models.Repo) error {
	stmt := sql.Lookup(db.driver, SQLRepoBatch)
	for _, repo := range repos {
		if _, err := db.Exec(stmt,
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