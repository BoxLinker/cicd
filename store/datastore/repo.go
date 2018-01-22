package datastore

import (
	"fmt"

	"github.com/BoxLinker/cicd/models"
	"github.com/BoxLinker/cicd/store/datastore/sql"
	"github.com/Sirupsen/logrus"
	"github.com/russross/meddler"
)

const repoTable = "repos"

func (db *datastore) GetRepo(id int64) (*models.Repo, error) {
	var repo = new(models.Repo)
	var err = meddler.Load(db, repoTable, repo, id)
	return repo, err
}

func (db *datastore) RepoList(u *models.User) []*models.Repo {
	stmt := sql.Lookup(db.driver, SQLQueryReposByUserID)
	data := make([]*models.Repo, 0)
	if err := meddler.QueryAll(db, &data, stmt, u.ID); err != nil {
		logrus.Errorf("DataStore RepoList err (%s)", err.Error())
		return nil
	}
	return data
}

func (db *datastore) UpdateRepo(repo *models.Repo) error {
	return meddler.Update(db, repoTable, repo)
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
			repo.Avatar,
			repo.Link,
			repo.Clone,
			repo.Branch,
			repo.Timeout,
			repo.IsPrivate,
			repo.IsTrusted,
			repo.IsActive,
			repo.AllowPull,
			repo.AllowPush,
			repo.AllowDeploy,
			repo.AllowTag,
			repo.Hash,
			repo.SCM,
			repo.Config,
			repo.IsGated,
			repo.Visibility,
			repo.Counter,
		); err != nil {
			return err
		}
	}
	return nil
}

func (db *datastore) GetRepoOwnerName(owner, repoName string) (*models.Repo, error) {
	stmt := sql.Lookup(db.driver, SQLRepoFindFullName)
	var repo = new(models.Repo)
	fullName := fmt.Sprintf("%s/%s", owner, repoName)
	logrus.Debugf("sql:> %s \n\tparams:> %s", rebind(stmt), fullName)
	var err = meddler.QueryRow(db, repo, rebind(stmt), fullName)
	if err != nil {
		logrus.Errorf("[DataStore] GetRepoOwnerName error: %s", err)
	}
	return repo, err
}
