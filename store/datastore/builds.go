package datastore

import (
	"github.com/BoxLinker/cicd/models"
	"fmt"
	"github.com/BoxLinker/cicd/store/datastore/sql"
	"time"
	"github.com/russross/meddler"
)

func (db *datastore) CreateBuild(build *models.Build, procs ...*models.Proc) error {
	id, err := db.incrementRepoRetry(build.RepoID)
	if err != nil {
		return err
	}
	build.Number = id
	build.Created = time.Now().UTC().Unix()
	build.Enqueued = build.Created
	err = meddler.Insert(db, "builds", build)
	if err != nil {
		return err
	}
	for _, proc := range procs {
		proc.BuildID = build.ID
		err = meddler.Insert(db, "procs", proc)
		if err != nil {
			return err
		}
	}
	return nil
}

func (db *datastore) UpdateBuild(build *models.Build) error {
	return meddler.Update(db, "builds", build)
}

func (db *datastore) incrementRepoRetry(id int64) (int, error) {
	repo, err := db.GetRepo(id)
	if err != nil {
		return 0, fmt.Errorf("database: cannot fetch repository: %s", err)
	}
	for i := 0; i < 10; i++ {
		seq, err := db.incrementRepo(repo.ID, repo.Counter+i, repo.Counter+i+1)
		if err != nil {
			return 0, err
		}
		if seq == 0 {
			continue
		}
		return seq, nil
	}
	return 0, fmt.Errorf("cannot increment next build number")
}

func (db *datastore) incrementRepo(id int64, old, new int) (int, error) {
	results, err := db.Exec(sql.Lookup(db.driver, "repo-update-counter"), new, old, id)
	if err != nil {
		return 0, fmt.Errorf("database: update repository counter: %s", err)
	}
	updated, err := results.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("database: update repository counter: %s", err)
	}
	if updated == 0 {
		return 0, nil
	}
	return new, nil
}