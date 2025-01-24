package models

import "time"

type ECommBill struct {
	BaseModel
	Name           string    `gorm:"type:varchar(100);not null" json:"name"`
	OrderNumber    string    `gorm:"type:varchar(100);not null" json:"orderNumber"`
	TrackingNumber string    `gorm:"type:varchar(100);not null" json:"trackingNumber"`
	Title          string    `gorm:"type:varchar(100);not null" json:"title"`
	Specification  string    `gorm:"type:varchar(100);not null" json:"specification"`
	Amount         int       `gorm:"type:int(11);not null" json:"amount"`
	DeliveryTime   time.Time `gorm:"type:Time;not null" json:"deliveryTime"`
}
