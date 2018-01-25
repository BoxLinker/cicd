package application

import (
	"github.com/go-xorm/xorm"
	"time"
)

type PodConfigure struct {
	Memory string `xorm:"pk NOT NULL"`
	CPU string `xorm:"pk NOT NULL"`

	Created     time.Time `xorm:"-"`
	CreatedUnix int64
	Updated     time.Time `xorm:"-"`
	UpdatedUnix int64
}

func (me *PodConfigure) BeforeInsert() {
	me.CreatedUnix = time.Now().Unix()
	me.UpdatedUnix = me.CreatedUnix
}

func (me *PodConfigure) BeforeUpdate() {
	me.UpdatedUnix = time.Now().Unix()
}

func (me *PodConfigure) AfterSet(colName string, _ xorm.Cell) {
	switch colName {
	case "created_unix":
		me.Created = time.Unix(me.CreatedUnix, 0).Local()
	case "updated_unix":
		me.Updated = time.Unix(me.UpdatedUnix, 0)
	}
}

func (me *PodConfigure) APIJson() map[string]interface{} {
	return map[string]interface{}{
		"memory": me.Memory,
		"cpu": me.CPU,
		"created": time.Unix(me.CreatedUnix, 0).Format("2006-01-02"),
		"updated": time.Unix(me.UpdatedUnix, 0).Format("2006-01-02"),
	}
}
