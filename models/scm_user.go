package models

import (
	"time"
)

type SCMUser struct {
	ID int64 `meddler:"user_id,pk" json:"id"`
	UCenterID string `meddler:"user_center_id" json:"u_center_id"`
	Login string `meddler:"user_login" json:"login"`
	Email string `meddler:"user_email" json:"email"`
	Token string `meddler:"-"`// boxlinker user auth token
	AccessToken string `meddler:"access_token" json:"-"` // vcs oauth2 token
	SCM string `meddler:"user_scm" json:"scm,omitempty"`

	Created     time.Time `meddler:"-"`
	CreatedUnix int64 `meddler:"created_unix"`
	Updated     time.Time `meddler:"-"`
	UpdatedUnix int64	`meddler:"updated_unix"`
}
