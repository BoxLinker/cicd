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
ORDER BY branch_name ASC
LIMIT ? OFFSET ?

-- name: branch-del-repo-id

DELETE FROM branches
WHERE branch_repo_id = ?