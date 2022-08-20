package model

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        int32          `gorm:"primarykey" json:"id"` //int32对应bigint
	CreatedAt time.Time      `gorm:"column:add_time" json:"-"`
	UpdatedAt time.Time      `gorm:"column:update_time" json:"-"`
	DeletedAt gorm.DeletedAt `json:"-"`
	IsDelete  bool           `json:"-"`
}
