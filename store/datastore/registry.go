package datastore

import (
	"github.com/russross/meddler"
	"github.com/BoxLinker/cicd/models"
	"github.com/BoxLinker/cicd/store/datastore/sql"
)

func (db *datastore) RegistryFind(repo *models.Repo, addr string) (*models.Registry, error) {
	stmt := sql.Lookup(db.driver, "registry-find-repo-addr")
	data := new(models.Registry)
	err := meddler.QueryRow(db, data, stmt, repo.ID, addr)
	return data, err
}

func (db *datastore) RegistryList(repo *models.Repo) ([]*models.Registry, error) {
	stmt := sql.Lookup(db.driver, "registry-find-repo")
	data := []*models.Registry{}
	err := meddler.QueryAll(db, &data, stmt, repo.ID)
	return data, err
}

func (db *datastore) RegistryCreate(registry *models.Registry) error {
	return meddler.Insert(db, "registry", registry)
}

func (db *datastore) RegistryUpdate(registry *models.Registry) error {
	return meddler.Update(db, "registry", registry)
}

func (db *datastore) RegistryDelete(registry *models.Registry) error {
	stmt := sql.Lookup(db.driver, "registry-delete")
	_, err := db.Exec(stmt, registry.ID)
	return err
}
