package mysql

// Lookup returns the named statement.
func Lookup(name string) string {
	return index[name]
}

var index = map[string]string{
	"branch-insert-ignore":      branchInsertIgnore,
	"branch-find-repo-id":       branchFindRepoId,
	"branch-del-repo-id":        branchDelRepoId,
	"config-find-id":            configFindId,
	"config-find-repo-hash":     configFindRepoHash,
	"config-find-approved":      configFindApproved,
	"files-find-build":          filesFindBuild,
	"files-find-proc-name":      filesFindProcName,
	"files-find-proc-name-data": filesFindProcNameData,
	"files-delete-build":        filesDeleteBuild,
	"logs-find-proc":            logsFindProc,
	"procs-find-id":             procsFindId,
	"procs-find-build":          procsFindBuild,
	"procs-find-build-pid":      procsFindBuildPid,
	"procs-find-build-ppid":     procsFindBuildPpid,
	"procs-delete-build":        procsDeleteBuild,
	"registry-find-repo":        registryFindRepo,
	"registry-find-repo-addr":   registryFindRepoAddr,
	"registry-delete-repo":      registryDeleteRepo,
	"registry-delete":           registryDelete,
	"repo-update-counter":       repoUpdateCounter,
	"repo-insert-ignore":        repoInsertIgnore,
	"repo-find-user":            repoFindUser,
	"repo-find-fullName":        repoFindFullName,
	"repo-del-id":               repoDelId,
	"secret-find-repo":          secretFindRepo,
	"secret-find-repo-name":     secretFindRepoName,
	"secret-delete":             secretDelete,
	"task-list":                 taskList,
	"task-delete":               taskDelete,
	"scm_user-find-u_center_id": scmuserFindUcenterid,
	"scm_user-update":           scmuserUpdate,
	"user-find-id-scm":          userFindIdScm,
}

var branchInsertIgnore = `
INSERT IGNORE INTO branches (
 branch_name
,branch_repo_id
) VALUES (?,?)
`

var branchFindRepoId = `
SELECT
 branch_name
,branch_repo_id
FROM branches
WHERE branch_repo_id = ?
LIMIT ? OFFSET ?
ORDER BY branch_name ASC
`

var branchDelRepoId = `
DELETE FROM branches
WHERE branch_repo_id = ?
`

var configFindId = `
SELECT
 config_id
,config_repo_id
,config_hash
,config_data
FROM config
WHERE config_id = ?
`

var configFindRepoHash = `
SELECT
 config_id
,config_repo_id
,config_hash
,config_data
FROM config
WHERE config_repo_id = ?
  AND config_hash    = ?
`

var configFindApproved = `
SELECT build_id FROM builds
WHERE build_repo_id = ?
AND build_config_id = ?
AND build_status NOT IN ('blocked', 'pending')
LIMIT 1
`

var filesFindBuild = `
SELECT
 file_id
,file_build_id
,file_proc_id
,file_pid
,file_name
,file_mime
,file_size
,file_time
,file_meta_passed
,file_meta_failed
,file_meta_skipped
FROM files
WHERE file_build_id = ?
`

var filesFindProcName = `
SELECT
 file_id
,file_build_id
,file_proc_id
,file_pid
,file_name
,file_mime
,file_size
,file_time
,file_meta_passed
,file_meta_failed
,file_meta_skipped
FROM files
WHERE file_proc_id = ?
  AND file_name    = ?
`

var filesFindProcNameData = `
SELECT
 file_id
,file_build_id
,file_proc_id
,file_pid
,file_name
,file_mime
,file_size
,file_time
,file_meta_passed
,file_meta_failed
,file_meta_skipped
,file_data
FROM files
WHERE file_proc_id = ?
  AND file_name    = ?
`

var filesDeleteBuild = `
DELETE FROM files WHERE file_build_id = ?
`

var logsFindProc = `
SELECT
 log_id
,log_job_id
,log_data
FROM logs
WHERE log_job_id = ?
LIMIT 1
`

var procsFindId = `
SELECT
 proc_id
,proc_build_id
,proc_pid
,proc_ppid
,proc_pgid
,proc_name
,proc_state
,proc_error
,proc_exit_code
,proc_started
,proc_stopped
,proc_machine
,proc_platform
,proc_environ
FROM procs
WHERE proc_id = ?
`

var procsFindBuild = `
SELECT
 proc_id
,proc_build_id
,proc_pid
,proc_ppid
,proc_pgid
,proc_name
,proc_state
,proc_error
,proc_exit_code
,proc_started
,proc_stopped
,proc_machine
,proc_platform
,proc_environ
FROM procs
WHERE proc_build_id = ?
ORDER BY proc_id ASC
`

var procsFindBuildPid = `
SELECT
proc_id
,proc_build_id
,proc_pid
,proc_ppid
,proc_pgid
,proc_name
,proc_state
,proc_error
,proc_exit_code
,proc_started
,proc_stopped
,proc_machine
,proc_platform
,proc_environ
FROM procs
WHERE proc_build_id = ?
  AND proc_pid = ?
`

var procsFindBuildPpid = `
SELECT
proc_id
,proc_build_id
,proc_pid
,proc_ppid
,proc_pgid
,proc_name
,proc_state
,proc_error
,proc_exit_code
,proc_started
,proc_stopped
,proc_machine
,proc_platform
,proc_environ
FROM procs
WHERE proc_build_id = ?
  AND proc_ppid = ?
  AND proc_name = ?
`

var procsDeleteBuild = `
DELETE FROM procs WHERE proc_build_id = ?
`

var registryFindRepo = `
SELECT
 registry_id
,registry_repo_id
,registry_addr
,registry_username
,registry_password
,registry_email
,registry_token
FROM registry
WHERE registry_repo_id = ?
`

var registryFindRepoAddr = `
SELECT
 registry_id
,registry_repo_id
,registry_addr
,registry_username
,registry_password
,registry_email
,registry_token
FROM registry
WHERE registry_repo_id = ?
  AND registry_addr = ?
`

var registryDeleteRepo = `
DELETE FROM registry WHERE registry_repo_id = ?
`

var registryDelete = `
DELETE FROM registry WHERE registry_id = ?
`

var repoUpdateCounter = `
UPDATE repos SET repo_counter = ?
WHERE repo_counter = ?
  AND repo_id = ?
`

var repoInsertIgnore = `
INSERT IGNORE INTO repos (
 repo_user_id
,repo_owner
,repo_name
,repo_full_name
,repo_avatar
,repo_link
,repo_clone
,repo_branch
,repo_timeout
,repo_private
,repo_trusted
,repo_active
,repo_allow_pr
,repo_allow_push
,repo_allow_deploys
,repo_allow_tags
,repo_hash
,repo_scm
,repo_config_path
,repo_gated
,repo_visibility
,repo_counter
) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
`

var repoFindUser = `
SELECT
 repo_id
,repo_user_id
,repo_owner
,repo_name
,repo_full_name
,repo_avatar
,repo_link
,repo_clone
,repo_branch
,repo_timeout
,repo_private
,repo_trusted
,repo_active
,repo_allow_pr
,repo_allow_push
,repo_allow_deploys
,repo_allow_tags
,repo_hash
,repo_scm
,repo_config_path
,repo_gated
,repo_visibility
,repo_counter
FROM repos
WHERE repo_user_id = ?
ORDER BY repo_name ASC
`

var repoFindFullName = `
SELECT * FROM repos
WHERE repo_full_name = ?
LIMIT 1
`

var repoDelId = `
DELETE FROM repos
WHERE repo_id = ?
`

var secretFindRepo = `
SELECT
 secret_id
,secret_repo_id
,secret_name
,secret_value
,secret_images
,secret_events
,secret_conceal
,secret_skip_verify
FROM secrets
WHERE secret_repo_id = ?
`

var secretFindRepoName = `
SELECT
secret_id
,secret_repo_id
,secret_name
,secret_value
,secret_images
,secret_events
,secret_conceal
,secret_skip_verify
FROM secrets
WHERE secret_repo_id = ?
  AND secret_name = ?
`

var secretDelete = `
DELETE FROM secrets WHERE secret_id = ?
`

var taskList = `
SELECT
 task_id
,task_data
,task_labels
FROM tasks
`

var taskDelete = `
DELETE FROM tasks WHERE task_id = ?
`

var scmuserFindUcenterid = `
SELECT
 user_id
,user_center_id
,user_scm
,user_login
,user_email
,access_token
,created_unix
,updated_unix
FROM users
WHERE user_center_id = ? AND user_scm = ?
LIMIT 1
`

var scmuserUpdate = `
UPDATE users
SET
,user_scm = ?
,user_login = ?
,user_email = ?
,access_token = ?
,updated_unix = ?
WHERE user_id = ?
`

var userFindIdScm = `
SELECT *
FROM users
WHERE user_id = ? AND user_scm = ?
LIMIT 1
`
