-- name: drop-repos-index-repoFullName

ALTER TABLE repos DROP INDEX repo_full_name;

-- name: create-repos-index-repoUnique

ALTER TABLE repos ADD CONSTRAINT repo_unique UNIQUE(repo_user_id,repo_full_name,repo_scm);