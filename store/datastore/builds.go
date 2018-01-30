package datastore

import (
	"fmt"
	"time"

	"github.com/BoxLinker/cicd/models"
	"github.com/BoxLinker/cicd/store/datastore/sql"
	"github.com/Sirupsen/logrus"
	"github.com/cabernety/gopkg/httplib"
	"github.com/russross/meddler"
)

func (db *datastore) GetBuild(id int64) (*models.Build, error) {
	var build = new(models.Build)
	var err = meddler.Load(db, "builds", build, id)
	return build, err
}

func (db *datastore) GetBuildNumber(repo *models.Repo, num int) (*models.Build, error) {
	var build = new(models.Build)
	sql := rebind(buildNumberQuery)
	logrus.Debugf("sql:GetBuildNumber:> %s, %d, %d", sql, repo.ID, num)
	var err = meddler.QueryRow(db, build, sql, repo.ID, num)
	if err != nil {
		logrus.Errorf("[DataStore:GetBuildNumber] error: %s", err)
	}
	return build, err
}

func (db *datastore) GetBuildLastBefore(repo *models.Repo, branch string, num int64) (*models.Build, error) {
	var build = new(models.Build)
	var err = meddler.QueryRow(db, build, rebind(buildLastBeforeQuery), repo.ID, branch, num)
	if err != nil {
		logrus.Errorf("[DataStore:GetBuildLastBefore] (repo:%d, branch:%s num:%d) error: %s", repo.ID, branch, num, err)
	}
	return build, err
}

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

func (db *datastore) SearchBuild(repo *models.Repo, search string, pc *httplib.PageConfig) []*models.Build {
	result := make([]*models.Build, 0)
	if err := meddler.QueryAll(db, &result, buildSearch, repo.ID, search+"%", pc.Limit(), pc.Offset()); err != nil {
		logrus.Errorf("SearchBuild err: %v", err)
		return nil
	}
	return result
}

func (db *datastore) QueryBranchBuild(repo *models.Repo, branch string) []*models.Build {
	result := make([]*models.Build, 0)
	if err := meddler.QueryAll(db, &result, build5BranchQuery, repo.ID, branch); err != nil {
		logrus.Errorf("QueryBranchBuild err: %v", err)
		return nil
	}
	return result
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

const buildNumberQuery = `
SELECT *
FROM builds
WHERE build_repo_id = ?
  AND build_number = ?
LIMIT 1;
`

const buildLastBeforeQuery = `
SELECT *
FROM builds
WHERE build_repo_id = ?
  AND build_branch = ?
  AND build_id < ?
ORDER BY build_number DESC
LIMIT 1
`
const build5BranchQuery = `
SELECT *
FROM builds
WHERE build_repo_id = ?
	AND build_branch = ?
ORDER BY build_number DESC
LIMIT 5
`

const buildSearch = `
SELECT *
FROM builds
WHERE build_repo_id = ?
	AND build_branch like ?
ORDER BY build_number DESC
LIMIT ? OFFSET ?
`
