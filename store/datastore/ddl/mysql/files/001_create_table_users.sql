-- name: create-table-users

CREATE TABLE IF NOT EXISTS users (
 user_id                  INTEGER PRIMARY KEY AUTO_INCREMENT
,user_center_id           VARCHAR(250) NOT NULL
,user_scm                 VARCHAR(250)
,user_login               VARCHAR(250) NOT NULL
,user_email               VARCHAR(250)
,access_token             VARCHAR(500) NOT NULL
,created_unix             INTEGER
,updated_unix             INTEGER

,UNIQUE (user_center_id,user_login,user_scm)
);