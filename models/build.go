package models

import (
	"time"
	"github.com/go-xorm/xorm"
	"github.com/satori/go.uuid"
)

type Build struct {
	ID string `xorm:"pk not null",json:"id"`

	Created     time.Time `xorm:"-"`
	CreatedUnix int64
	Updated     time.Time `xorm:"-"`
	UpdatedUnix int64
}

func (me *Build) BeforeInsert() {
	me.ID = uuid.NewV4().String()
	me.CreatedUnix = time.Now().Unix()
	me.UpdatedUnix = me.CreatedUnix
}

func (me *Build) BeforeUpdate() {
	me.UpdatedUnix = time.Now().Unix()
}

func (me *Build) AfterSet(colName string, _ xorm.Cell) {
	switch colName {
	case "created_unix":
		me.Created = time.Unix(me.CreatedUnix, 0).Local()
	case "updated_unix":
		me.Updated = time.Unix(me.UpdatedUnix, 0)
	}
}
