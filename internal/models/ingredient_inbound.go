package models

import "time"

type IngredientInBound struct {
	BaseModel
	IngredientID     *int                `gorm:"type:int(11)" json:"ingredientId"`
	Ingredient       *Ingredients        `gorm:"foreignKey:IngredientID" json:"ingredient"`
	Supplier         string              `gorm:"type:varchar(256); DEFAULT ''" json:"supplier"`
	Specification    string              `gorm:"type:varchar(256)" json:"specification"`
	Price            float64             `gorm:"type:decimal(12,2)" json:"price"`
	Balance          float64             `gorm:"type:decimal(16,4)" json:"balance"` // 余量
	Cost             float64             `gorm:"type:decimal(12,2)" json:"cost"`    // 成本
	TotalPrice       float64             `gorm:"type:decimal(12,2)" json:"totalPrice"`
	FinishPrice      float64             `gorm:"type:decimal(10,2)" json:"finishPrice"`
	FinishPriceStr   string              `gorm:"type:varchar(1024)" json:"finishPriceStr"`
	UnFinishPrice    float64             `gorm:"type:decimal(10,2)" json:"unFinishPrice"`
	Status           int                 `gorm:"type:int(11);not null" json:"status"` // 1:未完成支付 2:已支付
	StockNum         float64             `gorm:"type:decimal(16,4)" json:"stockNum"`
	StockUnit        int                 `gorm:"type:int(2)" json:"stockUnit"`
	StockUser        string              `gorm:"type:varchar(256)" json:"stockUser"`
	StockTime        time.Time           `gorm:"type:Time" json:"stockTime"`
	InAndOut         bool                `gorm:"type:tinyint(1)" json:"inAndOut"` // InAndOut True 入库 False 出库
	OperationType    string              `gorm:"type:varchar(256)" json:"operationType"`
	OperationDetails string              `gorm:"type:varchar(256)" json:"operationDetails"`
	FinishPriceList  []map[string]string `gorm:"-" json:"finishPriceList"`
}
