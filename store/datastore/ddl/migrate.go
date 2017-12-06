package ddl

import (
	"github.com/BoxLinker/cicd/store/datastore/ddlastore/ddl/mysql"
	"database/sql"
	"errors"
)

// Supported database drivers
const (
	DriverSqlite   = "sqlite3"
	DriverMysql    = "mysql"
	DriverPostgres = "postgres"
)

// Migrate performs the database migration. If the migration fails
// and error is returned.
func Migrate(driver string, db *sql.DB) error {
	switch driver {
	case DriverMysql:
		return mysql.Migrate(db)
	default:
		return errors.New("not implement default driver")
	}
}