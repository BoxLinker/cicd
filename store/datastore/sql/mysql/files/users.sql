-- name: scm_user-find-u_center_id

SELECT
 user_id
,user_center_id
,user_scm
,user_login
,user_email
,access_token
,created_unix
,updated_unix
FROM users
WHERE user_center_id = ? AND user_scm = ?
LIMIT 1

-- name: scm_user-update

UPDATE users
SET
,user_scm = ?
,user_login = ?
,user_email = ?
,access_token = ?
,updated_unix = ?
WHERE user_id = ?

-- name: user-find-id-scm

SELECT *
FROM users
WHERE user_id = ? AND user_scm = ?
LIMIT 1