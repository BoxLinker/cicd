package datastore

import (
	"github.com/BoxLinker/cicd/models"
	"github.com/BoxLinker/cicd/store/datastore/sql"
	"github.com/russross/meddler"
	"github.com/Sirupsen/logrus"
	"fmt"
)

func (db *datastore) GetRepo(id int64) (*models.Repo, error) {
	var repo = new(models.Repo)
	var err = meddler.Load(db, "repos", repo, id)
	return repo, err
}

func (db *datastore) RepoList(u *models.User) ([]*models.Repo) {
	stmt := sql.Lookup(db.driver, SQLQueryReposByUserID)
	data := make([]*models.Repo, 0)
	if err := meddler.QueryAll(db, &data, stmt, u.ID); err != nil {
		logrus.Errorf("DataStore RepoList err (%s)", err.Error())
		return nil
	}
	return data
}

func (db *datastore) RepoBatch(user *models.User, repos []*models.Repo) error {
	stmt := sql.Lookup(db.driver, SQLRepoBatch)
	for _, repo := range repos {
		repo.UserID = user.ID
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
		); err != nil {
			return err
		}
	}
	return nil
}

func (db *datastore) GetRepoOwnerName(owner, repoName string) (*models.Repo, error) {
	stmt := sql.Lookup(db.driver, SQLRepoFindFullName)
	var repo = new(models.Repo)
	var err = meddler.QueryRow(db, repo, rebind(stmt), fmt.Sprintf("%s/%s", owner, repoName))
	return repo, err
}