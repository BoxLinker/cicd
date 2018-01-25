package registry

import (
	"time"
	"github.com/go-xorm/xorm"
	"github.com/satori/go.uuid"
)

type Image struct {
	Id 		string 	`xorm:"pk"`
	Namespace 	string 	`xorm:"UNIQUE(image) NOT NULL"`
	Name 	string `xorm:"UNIQUE(image) NOT NULL"`
	Tag 	string 	`xorm:"UNIQUE(image) NOT NULL"`
	Size 	int64
	Digest 	string 	`xorm:"NOT NULL"`
	Description string `xorm:"VARCHAR(2000)"`
	HtmlDoc string `xorm:"LONGTEXT"`
	IsPrivate bool

	Created     time.Time `xorm:"-"`
	CreatedUnix int64
	Updated     time.Time `xorm:"-"`
	UpdatedUnix int64

}


func (me *Image) BeforeInsert() {
	me.Id = uuid.NewV4().String()
	me.CreatedUnix = time.Now().Unix()
	me.UpdatedUnix = me.CreatedUnix
}

func (me *Image) BeforeUpdate() {
	me.UpdatedUnix = time.Now().Unix()
}

func (me *Image) AfterSet(colName string, _ xorm.Cell) {
	switch colName {
	case "created_unix":
		me.Created = time.Unix(me.CreatedUnix, 0).Local()
	case "updated_unix":
		me.Updated = time.Unix(me.UpdatedUnix, 0)
	}
}

func (me *Image) APIJson() map[string]interface{} {
	return map[string]interface{}{
		"id": me.Id,
		"namespace": me.Namespace,
		"name": me.Name,
		"tag": me.Tag,
		"size": me.Size,
		"digest": me.Digest,
		"description": me.Description,
		"is_private": me.IsPrivate,
		"created": me.Created,
		"updated": me.Updated,
		"htmlDoc": me.HtmlDoc,
	}
}

func (me *Image) APISimpleJson() map[string]interface{} {
	return map[string]interface{}{
		"id": me.Id,
		"namespace": me.Namespace,
		"name": me.Name,
		"tag": me.Tag,
		"size": me.Size,
		"digest": me.Digest,
		"description": me.Description,
		"is_private": me.IsPrivate,
		"created": me.Created,
		"updated": me.Updated,
	}
}
