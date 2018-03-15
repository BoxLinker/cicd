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
		name: "create-table-builds",
		stmt: createTableBuilds,
	},
	{
		name: "create-index-builds-repo",
		stmt: createIndexBuildsRepo,
	},
	{
		name: "create-index-builds-author",
		stmt: createIndexBuildsAuthor,
	},
	{
		name: "create-table-config",
		stmt: createTableConfig,
	},
	{
		name: "create-table-tasks",
		stmt: createTableTasks,
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
	{
		name: "create-table-registry",
		stmt: createTableRegistry,
	},
	{
		name: "create-index-registry-repo",
		stmt: createIndexRegistryRepo,
	},
	{
		name: "create-table-secrets",
		stmt: createTableSecrets,
	},
	{
		name: "create-index-secrets-repo",
		stmt: createIndexSecretsRepo,
	},
	{
		name: "alter-table-add-repo-visibility",
		stmt: alterTableAddRepoVisibility,
	},
	{
		name: "update-table-set-repo-visibility",
		stmt: updateTableSetRepoVisibility,
	},
	{
		name: "alter-table-add-repo-seq",
		stmt: alterTableAddRepoSeq,
	},
	{
		name: "update-table-set-repo-seq",
		stmt: updateTableSetRepoSeq,
	},
	{
		name: "update-table-set-repo-seq-default",
		stmt: updateTableSetRepoSeqDefault,
	},
	{
		name: "alter-table-add-repo-active",
		stmt: alterTableAddRepoActive,
	},
	{
		name: "update-table-set-repo-active",
		stmt: updateTableSetRepoActive,
	},
	{
		name: "create-table-branches",
		stmt: createTableBranches,
	},
	{
		name: "drop-repos-index-repoFullName",
		stmt: dropReposIndexRepoFullName,
	},
	{
		name: "create-repos-index-repoUnique",
		stmt: createReposIndexRepoUnique,
	},
	{
		name: "drop-users-index",
		stmt: dropUsersIndex,
	},
	{
		name: "create_users-index",
		stmt: createusersIndex,
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

,UNIQUE (user_center_id,user_login,user_scm)
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
// 003_create_table_builds.sql
//

var createTableBuilds = `
CREATE TABLE IF NOT EXISTS builds (
 build_id        INTEGER PRIMARY KEY AUTO_INCREMENT
,build_repo_id   INTEGER
,build_number    INTEGER
,build_event     VARCHAR(500)
,build_status    VARCHAR(500)
,build_enqueued  INTEGER
,build_created   INTEGER
,build_started   INTEGER
,build_finished  INTEGER
,build_commit    VARCHAR(500)
,build_branch    VARCHAR(500)
,build_ref       VARCHAR(500)
,build_refspec   VARCHAR(1000)
,build_remote    VARCHAR(500)
,build_title     VARCHAR(1000)
,build_message   VARCHAR(2000)
,build_timestamp INTEGER
,build_author    VARCHAR(500)
,build_avatar    VARCHAR(1000)
,build_email     VARCHAR(500)
,build_link      VARCHAR(1000)
,build_deploy    VARCHAR(500)
,build_signed    BOOLEAN
,build_verified  BOOLEAN
,build_parent    INTEGER
,build_error     VARCHAR(500)
,build_reviewer  VARCHAR(250)
,build_reviewed  INTEGER
,build_sender    VARCHAR(250)
,build_config_id INTEGER

,UNIQUE(build_number, build_repo_id)
);
`

var createIndexBuildsRepo = `
CREATE INDEX ix_build_repo ON builds (build_repo_id);
`

var createIndexBuildsAuthor = `
CREATE INDEX ix_build_author ON builds (build_author);
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
// 005_create_table_tasks.sql
//

var createTableTasks = `
CREATE TABLE IF NOT EXISTS tasks (
 task_id      VARCHAR(250) PRIMARY KEY
,task_data    MEDIUMBLOB
,task_labels  MEDIUMBLOB
);
`

//
// 006_create_table_procs.sql
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
// 007_create_table_logs.sql
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
// 008_create_table_files.sql
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

//
// 009_create_table_registry.sql
//

var createTableRegistry = `
CREATE TABLE IF NOT EXISTS registry (
 registry_id        INTEGER PRIMARY KEY AUTO_INCREMENT
,registry_repo_id   INTEGER
,registry_addr      VARCHAR(250)
,registry_email     VARCHAR(500)
,registry_username  VARCHAR(2000)
,registry_password  VARCHAR(8000)
,registry_token     VARCHAR(2000)

,UNIQUE(registry_addr, registry_repo_id)
);
`

var createIndexRegistryRepo = `
CREATE INDEX ix_registry_repo ON registry (registry_repo_id);
`

//
// 010_create_table_secets.sql
//

var createTableSecrets = `
CREATE TABLE IF NOT EXISTS secrets (
 secret_id          INTEGER PRIMARY KEY AUTO_INCREMENT
,secret_repo_id     INTEGER
,secret_name        VARCHAR(250)
,secret_value       MEDIUMBLOB
,secret_images      VARCHAR(2000)
,secret_events      VARCHAR(2000)
,secret_skip_verify BOOLEAN
,secret_conceal     BOOLEAN

,UNIQUE(secret_name, secret_repo_id)
);
`

var createIndexSecretsRepo = `
CREATE INDEX ix_secrets_repo  ON secrets  (secret_repo_id);
`

//
// 011_add_column_repo_visibility.sql
//

var alterTableAddRepoVisibility = `
ALTER TABLE repos ADD COLUMN repo_visibility VARCHAR(50)
`

var updateTableSetRepoVisibility = `
UPDATE repos
SET repo_visibility = CASE
  WHEN repo_private = false THEN 'public'
  ELSE 'private'
  END
`

//
// 012_add_column_repo_seq.sql
//

var alterTableAddRepoSeq = `
ALTER TABLE repos ADD COLUMN repo_counter INTEGER;
`

var updateTableSetRepoSeq = `
UPDATE repos SET repo_counter = (
  SELECT max(build_number)
  FROM builds
  WHERE builds.build_repo_id = repos.repo_id
)
`

var updateTableSetRepoSeqDefault = `
UPDATE repos SET repo_counter = 0
WHERE repo_counter IS NULL
`

//
// 013_add_column_repo_active.sql
//

var alterTableAddRepoActive = `
ALTER TABLE repos ADD COLUMN repo_active BOOLEAN
`

var updateTableSetRepoActive = `
UPDATE repos SET repo_active = true
`

//
// 014_create_table_branches.sql
//

var createTableBranches = `
CREATE TABLE IF NOT EXISTS branches (
 branch_id            INTEGER PRIMARY KEY AUTO_INCREMENT
,branch_name       VARCHAR(250)
,branch_repo_id     INTEGER

,UNIQUE (branch_name,branch_repo_id)
);
`

//
// 015_alert_repos_index.sql
//

var dropReposIndexRepoFullName = `
ALTER TABLE repos DROP INDEX IF EXISTS repo_unique;
`

var createReposIndexRepoUnique = `
ALTER TABLE repos ADD CONSTRAINT repo_unique UNIQUE IF NOT EXISTS (repo_full_name,repo_scm);
`

var dropUsersIndex = `
ALTER TABLE users DROP INDEX IF EXISTS user_center_id;
`

var createusersIndex = `
ALTER TABLE users ADD CONSTRAINT users_unique UNIQUE IF NOT EXISTS (user_login,user_scm);
`
