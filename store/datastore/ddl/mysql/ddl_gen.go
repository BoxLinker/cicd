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
	{
		name: "create-table-tasks",
		stmt: createTableTasks,
	},
	{
		name: "create-table-config",
		stmt: createTableConfig,
	},
	{
		name: "create-table-procs",
		stmt: createTableProcs,
	},
	{
		name: "create-index-procs-build",
		stmt: createIndexProcsBuild,
	},
	{
		name: "create-table-logs",
		stmt: createTableLogs,
	},
	{
		name: "create-table-files",
		stmt: createTableFiles,
	},
	{
		name: "create-index-files-builds",
		stmt: createIndexFilesBuilds,
	},
	{
		name: "create-index-files-procs",
		stmt: createIndexFilesProcs,
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
CREATE TABLE IF NOT EXISTS users (
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
 repo_id            INTEGER PRIMARY KEY AUTO_INCREMENT
,repo_user_id       INTEGER
,repo_owner         VARCHAR(250)
,repo_name          VARCHAR(250)
,repo_full_name     VARCHAR(250)
,repo_avatar        VARCHAR(500)
,repo_link          VARCHAR(1000)
,repo_clone         VARCHAR(1000)
,repo_branch        VARCHAR(500)
,repo_timeout       INTEGER
,repo_private       BOOLEAN
,repo_trusted       BOOLEAN
,repo_allow_pr      BOOLEAN
,repo_allow_push    BOOLEAN
,repo_allow_deploys BOOLEAN
,repo_allow_tags    BOOLEAN
,repo_hash          VARCHAR(500)
,repo_scm           VARCHAR(50)
,repo_config_path   VARCHAR(500)
,repo_gated         BOOLEAN

,UNIQUE(repo_full_name)
);
`

//
// 003_create_table_tasks.sql
//

var createTableTasks = `
CREATE TABLE IF NOT EXISTS tasks (
 task_id      VARCHAR(250) PRIMARY KEY
,task_data    MEDIUMBLOB
,task_labels  MEDIUMBLOB
);
`

//
// 004_create_table_config.sql
//

var createTableConfig = `
CREATE TABLE IF NOT EXISTS config (
 config_id      INTEGER PRIMARY KEY AUTO_INCREMENT
,config_repo_id INTEGER
,config_hash    VARCHAR(250)
,config_data    MEDIUMBLOB

,UNIQUE(config_hash, config_repo_id)
);
`

//
// 005_create_table_procs.sql
//

var createTableProcs = `
CREATE TABLE IF NOT EXISTS procs (
 proc_id         INTEGER PRIMARY KEY AUTO_INCREMENT
,proc_build_id   INTEGER
,proc_pid        INTEGER
,proc_ppid       INTEGER
,proc_pgid       INTEGER
,proc_name       VARCHAR(250)
,proc_state      VARCHAR(250)
,proc_error      VARCHAR(500)
,proc_exit_code  INTEGER
,proc_started    INTEGER
,proc_stopped    INTEGER
,proc_machine    VARCHAR(250)
,proc_platform   VARCHAR(250)
,proc_environ    VARCHAR(2000)

,UNIQUE(proc_build_id, proc_pid)
);
`

var createIndexProcsBuild = `
CREATE INDEX proc_build_ix ON procs (proc_build_id);
`

//
// 006_create_table_logs.sql
//

var createTableLogs = `
CREATE TABLE IF NOT EXISTS logs (
 log_id     INTEGER PRIMARY KEY AUTO_INCREMENT
,log_job_id INTEGER
,log_data   MEDIUMBLOB

,UNIQUE(log_job_id)
);
`

//
// 007_create_table_files.sql
//

var createTableFiles = `
CREATE TABLE IF NOT EXISTS files (
 file_id       INTEGER PRIMARY KEY AUTO_INCREMENT
,file_build_id INTEGER
,file_proc_id  INTEGER
,file_name     VARCHAR(250)
,file_mime     VARCHAR(250)
,file_size     INTEGER
,file_time     INTEGER
,file_data     MEDIUMBLOB

,UNIQUE(file_proc_id,file_name)
);
`

var createIndexFilesBuilds = `
CREATE INDEX file_build_ix ON files (file_build_id);
`

var createIndexFilesProcs = `
CREATE INDEX file_proc_ix  ON files (file_proc_id);
`
