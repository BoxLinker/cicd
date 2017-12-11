package datastore

import (
	"github.com/BoxLinker/cicd/models"
	"io"
	"io/ioutil"
	"github.com/russross/meddler"
	"github.com/BoxLinker/cicd/store/datastore/sql"
	"bytes"
)

func (db *datastore) FileList(build *models.Build) ([]*models.File, error) {
	stmt := sql.Lookup(db.driver, "files-find-build")
	list := []*models.File{}
	err := meddler.QueryAll(db, &list, stmt, build.ID)
	return list, err
}

func (db *datastore) FileFind(proc *models.Proc, name string) (*models.File, error) {
	stmt := sql.Lookup(db.driver, "files-find-proc-name")
	file := new(models.File)
	err := meddler.QueryRow(db, file, stmt, proc.ID, name)
	return file, err
}

func (db *datastore) FileRead(proc *models.Proc, name string) (io.ReadCloser, error) {
	stmt := sql.Lookup(db.driver, "files-find-proc-name-data")
	file := new(fileData)
	err := meddler.QueryRow(db, file, stmt, proc.ID, name)
	buf := bytes.NewBuffer(file.Data)
	return ioutil.NopCloser(buf), err
}

func (db *datastore) FileCreate(file *models.File, r io.Reader) error {
	d, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	f := fileData{
		ID:      file.ID,
		BuildID: file.BuildID,
		ProcID:  file.ProcID,
		PID:     file.PID,
		Name:    file.Name,
		Size:    file.Size,
		Mime:    file.Mime,
		Time:    file.Time,
		Passed:  file.Passed,
		Failed:  file.Failed,
		Skipped: file.Skipped,
		Data:    d,
	}
	return meddler.Insert(db, "files", &f)
}

type fileData struct {
	ID      int64  `meddler:"file_id,pk"`
	BuildID int64  `meddler:"file_build_id"`
	ProcID  int64  `meddler:"file_proc_id"`
	PID     int    `meddler:"file_pid"`
	Name    string `meddler:"file_name"`
	Size    int    `meddler:"file_size"`
	Mime    string `meddler:"file_mime"`
	Time    int64  `meddler:"file_time"`
	Passed  int    `meddler:"file_meta_passed"`
	Failed  int    `meddler:"file_meta_failed"`
	Skipped int    `meddler:"file_meta_skipped"`
	Data    []byte `meddler:"file_data"`
}
