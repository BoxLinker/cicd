package datastore

import (
	"github.com/BoxLinker/cicd/models"
	"io"
	"github.com/BoxLinker/cicd/store/datastore/sql"
	"github.com/russross/meddler"
	"io/ioutil"
)

func (db *datastore) LogSave(proc *models.Proc, r io.Reader) error {
	stmt := sql.Lookup(db.driver, "logs-find-proc")
	data := new(logData)
	err := meddler.QueryRow(db, data, stmt, proc.ID)
	if err != nil {
		data = &logData{ProcID: proc.ID}
	}
	data.Data, _ = ioutil.ReadAll(r)
	return meddler.Save(db, "logs", data)
}

type logData struct {
	ID     int64  `meddler:"log_id,pk"`
	ProcID int64  `meddler:"log_job_id"`
	Data   []byte `meddler:"log_data"`
}