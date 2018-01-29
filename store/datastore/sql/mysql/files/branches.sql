-- name: branch-insert-ignore

INSERT IGNORE INTO branches (
 branch_name
,branch_repo_id
) VALUES (?,?)

-- name: branch-find-repo-id

SELECT
 branch_name
,branch_repo_id
FROM branches
WHERE branch_repo_id = ?
LIMIT ? OFFSET ?
ORDER BY branch_name ASC

-- name: branch-del-repo-id

DELETE FROM branches
WHERE branch_repo_id = ?