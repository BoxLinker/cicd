-- name: drop-repos-index-repoFullName

ALTER TABLE repos DROP INDEX IF EXISTS repo_unique;

-- name: create-repos-index-repoUnique

ALTER TABLE repos ADD CONSTRAINT repo_unique UNIQUE(repo_full_name,repo_scm);

-- name: drop-users-index

ALTER TABLE users DROP INDEX IF EXISTS user_center_id;

-- name: create_users-index

ALTER TABLE users ADD CONSTRAINT users_unique UNIQUE(user_login,user_scm);