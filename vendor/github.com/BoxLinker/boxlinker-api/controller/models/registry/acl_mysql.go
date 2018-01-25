package registry

import (
	"time"
	"github.com/go-xorm/xorm"
	"github.com/satori/go.uuid"
	"strings"
)

type ACL struct {
	Id string `json:"id" xorm:"pk UNIQUE NOT NULL"`
	Account string `json:"account,omitempty" xorm:"account"`
	Type string `json:"type,omitempty" xorm:"type"`
	Name string `json:"name,omitempty" xorm:"name"`
	IP string `json:"ip,omitempty" xorm:"ip"`
	Service string `json:"service,omitempty" xorm:"service"`
	Actions string `json:"actions,omitempty" xorm:"actions"`
	ActionsArray []string `xorm:"-"`
	Comment string `json:"comment,omitempty" xorm:"comment"`

	Created     time.Time `xorm:"-"`
	CreatedUnix int64
	Updated     time.Time `xorm:"-"`
	UpdatedUnix int64
}
func (me *ACL) BeforeInsert() {
	me.Id = uuid.NewV4().String()
	me.CreatedUnix = time.Now().Unix()
	me.UpdatedUnix = me.CreatedUnix
}

func (me *ACL) BeforeUpdate() {
	me.UpdatedUnix = time.Now().Unix()
}

func (me *ACL) AfterSet(colName string, _ xorm.Cell) {
	switch colName {
	case "created_unix":
		me.Created = time.Unix(me.CreatedUnix, 0).Local()
	case "updated_unix":
		me.Updated = time.Unix(me.UpdatedUnix, 0).Local()
	case "actions":
		me.ActionsArray = strings.Split(me.Actions,",")
	}
}
