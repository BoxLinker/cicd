package datastore

import (
	"fmt"
	"strings"

	"github.com/BoxLinker/cicd/models"
	"github.com/BoxLinker/cicd/store"
	"github.com/BoxLinker/cicd/store/datastore/sql"
	"github.com/Sirupsen/logrus"
	"github.com/russross/meddler"
)

const repoTable = "repos"

func (db *datastore) RepoCount(user *models.User) int {
	stmt := "select count(*) count from repos where repo_user_id = ?"
	var result struct {
		Count int `meddler:"count"`
	}
	if err := meddler.QueryRow(db, &result, stmt, user.ID); err != nil {
		logrus.Errorf("RepoCount error: %v", err)
		return 0
	}
	return result.Count
}

func (db *datastore) GetRepo(id int64) (*models.Repo, error) {
	var repo = new(models.Repo)
	var err = meddler.Load(db, repoTable, repo, id)
	return repo, err
}

func (db *datastore) RepoList(opt *store.RepoListOptions) []*models.Repo {
	stmt := "SELECT * FROM repos WHERE 1=1"
	args := make([]interface{}, 0)
	if opt.User != nil {
		stmt += " AND repo_user_id = ?"
		args = append(args, opt.User.ID)
	}
	if !opt.All {
		stmt += " AND repo_active = ?"
		if opt.Active {
			args = append(args, "1")
		} else {
			args = append(args, "0")
		}
	}
	stmt += " ORDER BY repo_name ASC"
	if opt.Pagination != nil {
		limit := opt.Pagination.Limit()
		offset := opt.Pagination.Offset()
		if limit >= 0 {
			stmt += " LIMIT ?"
			args = append(args, limit)
		}
		if offset >= 0 {
			stmt += " OFFSET ?"
			args = append(args, offset)
		}
	}
	data := make([]*models.Repo, 0)
	logrus.Infof("sql:> %s", fmt.Sprintf(strings.Replace(stmt, "?", "%v", -1), args...))
	if err := meddler.QueryAll(db, &data, stmt, args...); err != nil {
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

func (db *datastore) GetRepoOwnerName(owner, repoName, scm string) (*models.Repo, error) {
	stmt := sql.Lookup(db.driver, SQLRepoFindFullName)
	var repo = new(models.Repo)
	fullName := fmt.Sprintf("%s/%s", owner, repoName)
	logrus.Debugf("sql:> %s \n\tparams:> %s, %s", rebind(stmt), fullName, scm)
	var err = meddler.QueryRow(db, repo, rebind(stmt), fullName, scm)
	if err != nil {
		logrus.Errorf("[DataStore] GetRepoOwnerName error: %s", err)
	}
	return repo, err
}
