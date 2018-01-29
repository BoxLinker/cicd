-- name: create-table-branches

CREATE TABLE IF NOT EXISTS branches (
 branch_id            INTEGER PRIMARY KEY AUTO_INCREMENT
,branch_name       VARCHAR(250)
,branch_repo_id     INTEGER

,UNIQUE (branch_name,branch_repo_id)
);