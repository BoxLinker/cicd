package mysql

import (
	"database/sql"
)

var migrations = []struct {
	name string
	stmt string
}{
	{
		name: "create-table-users",
		stmt: createTableUsers,
	},
	{
		name: "create-table-repos",
		stmt: createTableRepos,
	},
}

// Migrate performs the database migration. If the migration fails
// and error is returned.
func Migrate(db *sql.DB) error {
	if err := createTable(db); err != nil {
		return err
	}
	completed, err := selectCompleted(db)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	for _, migration := range migrations {
		if _, ok := completed[migration.name]; ok {

			continue
		}

		if _, err := db.Exec(migration.stmt); err != nil {
			return err
		}
		if err := insertMigration(db, migration.name); err != nil {
			return err
		}

	}
	return nil
}

func createTable(db *sql.DB) error {
	_, err := db.Exec(migrationTableCreate)
	return err
}

func insertMigration(db *sql.DB, name string) error {
	_, err := db.Exec(migrationInsert, name)
	return err
}

func selectCompleted(db *sql.DB) (map[string]struct{}, error) {
	migrations := map[string]struct{}{}
	rows, err := db.Query(migrationSelect)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		migrations[name] = struct{}{}
	}
	return migrations, nil
}

//
// migration table ddl and sql
//

var migrationTableCreate = `
CREATE TABLE IF NOT EXISTS migrations (
 name VARCHAR(255)
,UNIQUE(name)
)
`

var migrationInsert = `
INSERT INTO migrations (name) VALUES (?)
`

var migrationSelect = `
SELECT name FROM migrations
`

//
// 001_create_table_users.sql
//

var createTableUsers = `
CREATE TABLE IF NOT EXISTS scm_users (
 user_id                  INTEGER PRIMARY KEY AUTO_INCREMENT
,user_center_id           VARCHAR(250) NOT NULL
,user_scm                 VARCHAR(250)
,user_login               VARCHAR(250) NOT NULL
,user_email               VARCHAR(250)
,access_token             VARCHAR(500) NOT NULL
,created_unix             INTEGER
,updated_unix             INTEGER

,UNIQUE (user_login,user_scm)
);
`

//
// 002_create_table_repos.sql
//

var createTableRepos = `
CREATE TABLE IF NOT EXISTS repos (
 repo_id              INTEGER PRIMARY KEY AUTO_INCREMENT
,repo_user_id         VARCHAR(250) NOT NULL
,repo_owner           VARCHAR(250) NOT NULL
,repo_name            VARCHAR(250) NOT NULL
,repo_full_name       VARCHAR(250) NOT NULL
,repo_scm             VARCHAR(250)
,repo_link_url        VARCHAR(250)
,repo_clone_Url       VARCHAR(250)
,repo_default_branch  VARCHAR(250)
,repo_is_private      BOOLEAN
,created_unix         INTEGER
,updated_unix         INTEGER

,UNIQUE (repo_full_name,repo_scm)
);
`
