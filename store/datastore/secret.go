package datastore

import (
	"github.com/russross/meddler"
	"github.com/BoxLinker/cicd/models"
	"github.com/BoxLinker/cicd/store/datastore/sql"
)

func (db *datastore) SecretFind(repo *models.Repo, name string) (*models.Secret, error) {
	stmt := sql.Lookup(db.driver, "secret-find-repo-name")
	data := new(models.Secret)
	err := meddler.QueryRow(db, data, stmt, repo.ID, name)
	return data, err
}

func (db *datastore) SecretList(repo *models.Repo) ([]*models.Secret, error) {
	stmt := sql.Lookup(db.driver, "secret-find-repo")
	data := []*models.Secret{}
	err := meddler.QueryAll(db, &data, stmt, repo.ID)
	return data, err
}

func (db *datastore) SecretCreate(secret *models.Secret) error {
	return meddler.Insert(db, "secrets", secret)
}

func (db *datastore) SecretUpdate(secret *models.Secret) error {
	return meddler.Update(db, "secrets", secret)
}

func (db *datastore) SecretDelete(secret *models.Secret) error {
	stmt := sql.Lookup(db.driver, "secret-delete")
	_, err := db.Exec(stmt, secret.ID)
	return err
}
