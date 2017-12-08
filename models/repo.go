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
	Timeout     int64  `json:"timeout,omitempty"        meddler:"repo_timeout"`

	IsGated     bool   `json:"gated"                    meddler:"repo_gated"`

	IsActive    bool   `json:"active"                   meddler:"repo_active"`
	AllowPull   bool   `json:"allow_pr"                 meddler:"repo_allow_pr"`
	AllowPush   bool   `json:"allow_push"               meddler:"repo_allow_push"`
	AllowDeploy bool   `json:"allow_deploys"            meddler:"repo_allow_deploys"`
	AllowTag    bool   `json:"allow_tags"               meddler:"repo_allow_tags"`

	Counter     int    `json:"last_build"               meddler:"repo_counter"`
	Config      string `json:"config_file"              meddler:"repo_config_path"`

	Created     time.Time `json:"created"           meddler:"-"`
	CreatedUnix int64 	`json:"-"                   meddler:"created_unix"`
	Updated     time.Time `json:"updated"           meddler:"-"`
	UpdatedUnix int64	`json:"-"                   meddler:"updated_unix"`

}
