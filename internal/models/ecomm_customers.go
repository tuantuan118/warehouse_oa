package models

type ECommCustomers struct {
	BaseModel
	Name     string `gorm:"type:varchar(100);not null" json:"name"`
	ShopName string `gorm:"type:varchar(100);not null" json:"shopName"`
}
