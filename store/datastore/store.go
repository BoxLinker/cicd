package datastore

import (
	"database/sql"
	"time"

	"github.com/BoxLinker/cicd/store"
	"github.com/BoxLinker/cicd/store/datastore/ddl"
	"github.com/Sirupsen/logrus"
	_ "github.com/go-sql-driver/mysql"
	"github.com/russross/meddler"
)

const (
	TableSCMUsers = "users"
	TableRepos    = "repos"
)

const (
	SQLSCMUsersFindByUCenterID = "scm_user-find-u_center_id"
	SQLUserFindByIDSCM         = "user-find-id-scm"

	SQLQueryReposByUserID = "repo-find-user"
	SQLRepoBatch          = "repo-insert-ignore"
	SQLRepoFindFullName   = "repo-find-fullName"
	SQLRepoDelID          = "repo-del-id"
)

type datastore struct {
	*sql.DB
	driver string
}

func New(driver, datasource string) store.Store {
	return &datastore{
		DB:     open(driver, datasource),
		driver: driver,
	}
}

func open(driver, dataSource string) *sql.DB {
	logrus.Debugf("connect to database driver(%s) dataSource(%s)", driver, dataSource)
	db, err := sql.Open(driver, dataSource)
	if err != nil {
		logrus.Errorln(err)
		logrus.Fatalln("database connection failed")
	}
	if driver == "mysql" {
		// per issue https://github.com/go-sql-driver/mysql/issues/257
		db.SetMaxIdleConns(0)
	}

	setupMeddler(driver)

	if err := pingDatabase(db); err != nil {
		logrus.Errorln(err)
		logrus.Fatalln("database ping attempts failed")
	}

	logrus.Infoln("db connection ok!")

	if err := setupDatabase(driver, db); err != nil {
		logrus.Errorln(err)
		logrus.Fatalln("migration failed")
	}

	return db

}

// helper function to setup the databsae by performing
// automated database migration steps.
func setupDatabase(driver string, db *sql.DB) error {
	return ddl.Migrate(driver, db)
}

// helper function to ping the database with backoff to ensure
// a connection can be established before we proceed with the
// database setup and migration.
func pingDatabase(db *sql.DB) (err error) {
	for i := 0; i < 30; i++ {
		err = db.Ping()
		if err == nil {
			return
		}
		logrus.Infof("database ping failed. retry in 1s")
		time.Sleep(time.Second)
	}
	return
}

// helper function to setup the meddler default driver
// based on the selected driver name.
func setupMeddler(driver string) {
	switch driver {
	case "sqlite3":
		meddler.Default = meddler.SQLite
	case "mysql":
		meddler.Default = meddler.MySQL
	case "postgres":
		meddler.Default = meddler.PostgreSQL
	}
}
