package models

type FastBill struct {
	BaseModel
	OrderNumber    string  `gorm:"type:varchar(100);not null" json:"orderNumber"`
	TrackingNumber string  `gorm:"type:varchar(100);not null" json:"trackingNumber"`
	Title          string  `gorm:"type:varchar(100);not null" json:"title"`
	Specification  string  `gorm:"type:varchar(100);not null" json:"specification"`
	Amount         int     `gorm:"type:int(11);not null" json:"amount"`
	Status         int     `gorm:"type:int(2);not null" json:"status"`
	PayAmount      float64 `gorm:"type:decimal(10,2)" json:"payAmount"`
}
