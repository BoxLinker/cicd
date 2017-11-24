package mysql

// Lookup returns the named statement.
func Lookup(name string) string {
	return index[name]
}

var index = map[string]string{
	"repo-insert-ignore":        repoInsertIgnore,
	"repo-find-user":            repoFindUser,
	"scm_user-find-u_center_id": scmuserFindUcenterid,
	"scm_user-update":           scmuserUpdate,
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
FROM scm_users
WHERE user_center_id = ? AND user_scm = ?
LIMIT 1
`

var scmuserUpdate = `
UPDATE scm_users
SET
,user_scm = ?
,user_login = ?
,user_email = ?
,access_token = ?
,updated_unix = ?
WHERE user_id = ?
`
