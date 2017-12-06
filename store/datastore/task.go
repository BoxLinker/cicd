package datastore

import (
	"github.com/BoxLinker/cicd/models"
	"github.com/BoxLinker/cicd/store/datastore/sql"
	"github.com/russross/meddler"
)

func (db *datastore) TaskList() ([]*models.Task, error) {
	stmt := sql.Lookup(db.driver, "task-list")
	data := make([]*models.Task, 0)
	err := meddler.QueryAll(db, &data, stmt)
	return data, err
}

func (db *datastore) TaskInsert(task *models.Task) error {
	return meddler.Insert(db, "tasks", task)
}

func (db *datastore) TaskDelete(id string) error {
	stmt := sql.Lookup(db.driver, "task-delete")
	_, err := db.Exec(stmt, id)
	return err
}