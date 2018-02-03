package datastore

import (
	"fmt"
	"strings"
)

// todo
// rebind is a helper function that changes the sql
// bind type from ? to $ for postgres queries.
func rebind(query string) string {
	return query
}

func formatSql(stmt string, args ...interface{}) string {
	return fmt.Sprintf(strings.Replace(stmt, "?", "%v", -1), args...)
}
