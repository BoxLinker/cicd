package datastore

import (
	"github.com/BoxLinker/cicd/models"
	"github.com/BoxLinker/cicd/store/datastore/sql"
	"github.com/Sirupsen/logrus"
	"github.com/russross/meddler"
)

func (db *datastore) BranchList(repo *models.Repo, limit, offset int) []*models.Branch {
	stmt := sql.Lookup(db.driver, SQLBranchQueryRepoID)
	result := make([]*models.Branch, 0)
	if err := meddler.QueryAll(db, &result, stmt, repo.ID, limit, offset); err != nil {
		logrus.Errorf("BranchList err: %v", err)
		return nil
	}
	return result
}

func (db *datastore) BranchBatch(repo *models.Repo, branches []*models.Branch) error {
	stmt := sql.Lookup(db.driver, SQLBranchBatch)
	for _, branch := range branches {
		branch.RepoID = repo.ID
		if _, err := db.Exec(stmt,
			branch.Name,
			branch.RepoID,
		); err != nil {
			return err
		}
	}
	return nil
}
