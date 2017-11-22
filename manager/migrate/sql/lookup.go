package sql

// Supported database drivers
const (
	DriverSqlite   = "sqlite3"
	DriverMysql    = "mysql"
	DriverPostgres = "postgres"
)

func Lookup(driver string, name string) string {
	switch driver {
	default:

	}
}

