
-- name: repo-insert-or-update

INSERT INTO repos {
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
} VALUES (?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE repo_owner = '';