package mysql

// Lookup returns the named statement.
func Lookup(name string) string {
	return index[name]
}

var index = map[string]string{
	"repo-insert-ignore":        repoInsertIgnore,
	"repo-find-user":            repoFindUser,
	"repo-find-fullName":        repoFindFullName,
	"repo-del-id":               repoDelId,
	"task-list":                 taskList,
	"task-delete":               taskDelete,
	"scm_user-find-u_center_id": scmuserFindUcenterid,
	"scm_user-update":           scmuserUpdate,
	"user-find-id-scm":          userFindIdScm,
}

var repoInsertIgnore = `
INSERT IGNORE INTO repos (
  repo_user_id,
  repo_owner,
  repo_name,
  repo_full_name,
  repo_scm,
  repo_link_url,
  repo_clone_url,
  repo_default_branch,
  repo_is_private,
  created_unix,
  updated_unix
) VALUES (?,?,?,?,?,?,?,?,?,?,?)
`

var repoFindUser = `
SELECT
 repo_id
,repo_user_id
,repo_owner
,repo_name
,repo_full_name
,repo_scm
,repo_link_url
,repo_clone_url
,repo_default_branch
,repo_is_private
,created_unix
,updated_unix
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
