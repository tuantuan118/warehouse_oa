package models

type Customer struct {
	BaseModel
	Name     string `gorm:"type:varchar(256);not null" json:"name"`
	Address  string `gorm:"type:varchar(256);not null" json:"address"`
	Phone    string `gorm:"type:varchar(256);not null" json:"phone"`
	Email    string `gorm:"type:varchar(256);not null" json:"email"`
	Salesman string `gorm:"type:varchar(256);not null" json:"salesman"` // 销售人员
}
