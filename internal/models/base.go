package models

import (
	"time"
)

type BaseModel struct {
	ID        int       `gorm:"primaryKey" json:"id"`
	Operator  string    `gorm:"type:varchar(100)" json:"operator"`
	Remark    string    `gorm:"type:varchar(256)" json:"remark"`
	CreatedAt time.Time `gorm:"column:add_time" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:update_time" json:"updatedAt"`
}
