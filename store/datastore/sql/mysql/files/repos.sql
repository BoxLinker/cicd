-- name: repo-update-counter

UPDATE repos SET repo_counter = ?
WHERE repo_counter = ?
  AND repo_id = ?

-- name: repo-insert-ignore

INSERT IGNORE INTO repos (
 repo_user_id
,repo_owner
,repo_name
,repo_full_name
,repo_scm
,repo_link
,repo_clone
,repo_branch
,repo_private
) VALUES (?,?,?,?,?,?,?,?,?)

-- name: repo-find-user

SELECT
 repo_id
,repo_user_id
,repo_owner
,repo_name
,repo_full_name
,repo_scm
,repo_link
,repo_clone
,repo_branch
,repo_private
FROM repos
WHERE repo_user_id = ?
ORDER BY repo_name ASC

-- name: repo-find-fullName

SELECT * FROM repos
WHERE repo_full_name = ?
LIMIT 1

-- name: repo-del-id

DELETE FROM repos
WHERE repo_id = ?