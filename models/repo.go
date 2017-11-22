package models

import (
	"time"
	"github.com/go-xorm/xorm"
	"github.com/satori/go.uuid"
)

type Repo struct {
	ID		string 		`json:"repo_id" xorm:"pk NOT NULL"`
	UserID 	string 		`json:"-" xorm:"NOT NULL"`
	Owner 	string 		`json:"repo_owner"`
	Name 	string 		`json:"repo_name"`
	FullName string 	`json:"repo_full_name"`
	Kind 	string 		`json:"repo_scm"`
	Link 	string 		`json:"repo_link_url"`
	Clone 	string 		`json:"repo_clone_url"`
	Branch 	string		`json:"repo_default_branch"`
	IsPrivate bool 		`json:"repo_is_private"`

	Created     time.Time `xorm:"-"`
	CreatedUnix int64
	Updated     time.Time `xorm:"-"`
	UpdatedUnix int64
}


func (me *Repo) BeforeInsert() {
	me.ID = uuid.NewV4().String()
	me.CreatedUnix = time.Now().Unix()
	me.UpdatedUnix = me.CreatedUnix
}

func (me *Repo) BeforeUpdate() {
	me.UpdatedUnix = time.Now().Unix()
}

func (me *Repo) AfterSet(colName string, _ xorm.Cell) {
	switch colName {
	case "created_unix":
		me.Created = time.Unix(me.CreatedUnix, 0).Local()
	case "updated_unix":
		me.Updated = time.Unix(me.UpdatedUnix, 0)
	}
}

