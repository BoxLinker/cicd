package models

import (
	"time"
	"github.com/go-xorm/xorm"
	"github.com/satori/go.uuid"
)

func RollingUpdateTables() []interface{} {
	var tables []interface{}
	tables = append(tables, new(RollingUpdate))
	return tables
}

type RollingUpdate struct {
	Id 	string `xorm:"pk"`

	Created     time.Time `xorm:"-"`
	CreatedUnix int64
	Updated     time.Time `xorm:"-"`
	UpdatedUnix int64
}


func (me *RollingUpdate) BeforeInsert() {
	me.Id = uuid.NewV4().String()
	me.CreatedUnix = time.Now().Unix()
	me.UpdatedUnix = me.CreatedUnix
}

func (me *RollingUpdate) BeforeUpdate() {
	me.UpdatedUnix = time.Now().Unix()
}

func (me *RollingUpdate) AfterSet(colName string, _ xorm.Cell) {
	switch colName {
	case "created_unix":
		me.Created = time.Unix(me.CreatedUnix, 0).Local()
	case "updated_unix":
		me.Updated = time.Unix(me.UpdatedUnix, 0)
	}
}
