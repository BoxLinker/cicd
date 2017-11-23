-- name: create-table-repos

CREATE TABLE IF NOT EXISTS repos (
 repo_id              INTEGER PRIMARY KEY AUTO_INCREMENT
,repo_user_id         INTEGER
,repo_owner           VARCHAR(250)
,repo_name            VARCHAR(250)
,repo_full_name       VARCHAR(250)
,repo_scm             VARCHAR(250)
,repo_link_url        VARCHAR(250)
,repo_clone_Url       VARCHAR(250)
,repo_default_branch  VARCHAR(250)
,repo_is_private      BOOLEAN
,created_unix         INTEGER
,updated_unix         INTEGER

,UNIQUE (repo_full_name,repo_scm)
);