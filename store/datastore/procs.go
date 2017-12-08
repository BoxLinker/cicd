package datastore

import (
	"github.com/BoxLinker/cicd/models"
	"github.com/russross/meddler"
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
