package datastore

import (
	"github.com/BoxLinker/cicd/models"
	"github.com/russross/meddler"
	"github.com/BoxLinker/cicd/store/datastore/sql"
)

// todo 事务处理?
func (db *datastore) ProcCreate(procs []*models.Proc) error {
	for _, proc := range procs {
		if err := meddler.Insert(db, "procs", proc); err != nil {
			return err
		}
	}
	return nil
}

func (db *datastore) ProcLoad(id int64) (*models.Proc, error) {
	stmt := sql.Lookup(db.driver, "procs-find-id")
	proc := new(models.Proc)
	err := meddler.QueryRow(db, proc, stmt, id)
	return proc, err
}

func (db *datastore) ProcChild(build *models.Build, pid int, child string) (*models.Proc, error) {
	stmt := sql.Lookup(db.driver, "procs-find-build-ppid")
	proc := new(models.Proc)
	err := meddler.QueryRow(db, proc, stmt, build.ID, pid, child)
	return proc, err
}

func (db *datastore) ProcUpdate(proc *models.Proc) error {
	return meddler.Update(db, "procs", proc)
}

func (db *datastore) ProcList(build *models.Build) ([]*models.Proc, error) {
	stmt := sql.Lookup(db.driver, "procs-find-build")
	list := []*models.Proc{}
	err := meddler.QueryAll(db, &list, stmt, build.ID)
	return list, err
}