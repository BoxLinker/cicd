package sql

import "github.com/BoxLinker/cicd/store/datastore/sql/mysql"

// Supported database drivers
const (
	DriverSqlite   = "sqlite3"
	DriverMysql    = "mysql"
	DriverPostgres = "postgres"
)

func Lookup(driver string, name string) string {
	switch driver {
	case DriverMysql:
		return mysql.Lookup(name)
	default:
		return ""
	}
}

