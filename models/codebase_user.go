package models

import (
	"time"
	"github.com/go-xorm/xorm"
	"github.com/satori/go.uuid"
)

type CodeBaseUser struct {
	ID string `xorm:"pk NOT NULL" json:"id"`
	UserID string `xorm:"NOT NULL" json:"user_id"`
	Login string `json:"login"`
	Email string `json:"email"`
	Token string `xorm:"-"`// boxlinker user auth token
	AccessToken string `xorm:"NOT NULL" json:"access_token"` // vcs oauth2 token
	Kind string `json:"scm,omitempty"`

	Created     time.Time `xorm:"-"`
	CreatedUnix int64
	Updated     time.Time `xorm:"-"`
	UpdatedUnix int64
}

func (me *CodeBaseUser) BeforeInsert() {
	me.ID = uuid.NewV4().String()
	me.CreatedUnix = time.Now().Unix()
	me.UpdatedUnix = me.CreatedUnix
}

func (me *CodeBaseUser) BeforeUpdate() {
	me.UpdatedUnix = time.Now().Unix()
}

func (me *CodeBaseUser) AfterSet(colName string, _ xorm.Cell) {
	switch colName {
	case "created_unix":
		me.Created = time.Unix(me.CreatedUnix, 0).Local()
	case "updated_unix":
		me.Updated = time.Unix(me.UpdatedUnix, 0)
	}
}

