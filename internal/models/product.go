package models

type Product struct {
	BaseModel
	Name           string           `gorm:"uniqueIndex:idx_name_specification;type:varchar(256)" json:"name"`
	Specification  string           `gorm:"uniqueIndex:idx_name_specification;type:varchar(256)" json:"specification"`
	ProductContent []ProductContent `gorm:"foreignKey:ProductId;references:ID" json:"productContent"`
}

type ProductContent struct {
	ProductId  int       `gorm:"primaryKey;index" json:"productId"`
	FinishedId int       `gorm:"primaryKey;type:int(11)" json:"finishedId"`
	Finished   *Finished `gorm:"foreignKey:FinishedId;" json:"finished"`
	Quantity   float64   `gorm:"type:decimal(10,4);not null" json:"quantity"` // 用量
}

type ProductInventory struct {
	BaseModel
	ProductId     int      `gorm:"primaryKey;index" json:"productId"`
	Product       *Product `gorm:"foreignKey:ProductId;" json:"product"`
	Amount        int      `gorm:"type:int(11);not null" json:"amount"`
	ProductIdList string   `gorm:"-" json:"productIdList" form:"productIdList"`

	InventoryContent []InventoryContent `gorm:"foreignKey:InventoryId;" json:"inventoryContent"`
	// 记录产品使用的成品ID和数量

}

type InventoryContent struct {
	InventoryId int     `gorm:"primaryKey;index" json:"inventoryId"`
	FinishedId  int     `gorm:"primaryKey;type:int(11)" json:"finishedId"`
	Quantity    float64 `gorm:"type:decimal(10,4);not null" json:"quantity"` // 用量
}
