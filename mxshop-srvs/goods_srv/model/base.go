package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type GormList []string

type BaseModel struct {
	ID        int32          `gorm:"primarykey" json:"id"` //int32对应bigint
	CreatedAt time.Time      `gorm:"column:add_time" json:"-"`
	UpdatedAt time.Time      `gorm:"column:update_time" json:"-"`
	DeletedAt gorm.DeletedAt `json:"-"`
	IsDelete  bool           `json:"-"`
}

//实现sql.Scanner接口,Scan将value扫描至jsonb

func (g *GormList) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), &g)

}
func (g GormList) Value() (driver.Value, error) {
	return json.Marshal(g)
}
