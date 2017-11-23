package models

import (
	"time"
)

type Build struct {
	ID string `xorm:"pk not null",json:"id"`

	Created     time.Time `xorm:"-"`
	CreatedUnix int64
	Updated     time.Time `xorm:"-"`
	UpdatedUnix int64
}
