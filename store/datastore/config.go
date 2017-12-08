package datastore

import (
	gosql "database/sql"
	"github.com/BoxLinker/cicd/models"
	"github.com/BoxLinker/cicd/store/datastore/sql"
	"github.com/russross/meddler"
)

func (db *datastore) ConfigLoad(id int64) (*models.Config, error) {
	stmt := sql.Lookup(db.driver, "config-find-id")
	conf := new(models.Config)
	err := meddler.QueryRow(db, conf, stmt, id)
	return conf, err
}

func (db *datastore) ConfigFind(repo *models.Repo, hash string) (*models.Config, error) {
	stmt := sql.Lookup(db.driver, "config-find-repo-hash")
	conf := new(models.Config)
	err := meddler.QueryRow(db, conf, stmt, repo.ID, hash)
	return conf, err
}

func (db *datastore) ConfigFindApproved(config *models.Config) (bool, error) {
	var dest int64
	stmt := sql.Lookup(db.driver, "config-find-approved")
	err := db.DB.QueryRow(stmt, config.RepoID, config.ID).Scan(&dest)
	if err == gosql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (db *datastore) ConfigCreate(config *models.Config) error {
	return meddler.Insert(db, "config", config)
}