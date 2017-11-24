package store

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/Sirupsen/logrus"
	"time"
	"github.com/russross/meddler"
	"github.com/BoxLinker/cicd/store/ddl"
)

const (
	TableSCMUsers = "scm_users"
	TableRepos = "repos"
)

const (
	SQLSCMUsersFindByUCenterID = "scm_user-find-u_center_id"
	SQLQueryReposByUserID = "repo-find-user"
	SQLRepoBatch = "repo-insert-ignore"
)

type DataStore struct {
	*sql.DB
	driver string
}

func New(driver, datasource string) *DataStore {
	return &DataStore{
		DB: open(driver, datasource),
		driver: driver,
	}
}

func open(driver, dataSource string) *sql.DB {
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
