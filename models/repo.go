package models

import (
	"time"
)

type Repo struct {
	ID		int64 		`json:"repo_id"             meddler:"repo_id,pk"`
	UserID 	int64 		`json:"-"                   meddler:"repo_user_id"`
	Owner 	string 		`json:"owner"               meddler:"repo_owner"`
	Name 	string 		`json:"name"                meddler:"repo_name"`
	FullName string 	`json:"full_name"           meddler:"repo_full_name"`
	SCM 	string 		`json:"scm"                 meddler:"repo_scm"`
	Link 	string 		`json:"link_url"            meddler:"repo_link_url"`
	Clone 	string 		`json:"clone_url"           meddler:"repo_clone_url"`
	Branch 	string		`json:"default_branch"      meddler:"repo_default_branch"`
	IsPrivate bool 		`json:"is_private"          meddler:"repo_is_private"`

	Created     time.Time `json:"created"           meddler:"-"`
	CreatedUnix int64 	`json:"-"                   meddler:"created_unix"`
	Updated     time.Time `json:"updated"           meddler:"-"`
	UpdatedUnix int64	`json:"-"                   meddler:"updated_Unix"`
}
