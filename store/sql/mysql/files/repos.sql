
-- name: repo-insert-ignore

INSERT IGNORE INTO repos {
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
} VALUES (?,?,?,?,?,?,?,?,?,?)

-- name: repo-find-user

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
,updated_nix
FROM repos
WHERE repo_user_id = ?
ORDER BY repo_name ASC