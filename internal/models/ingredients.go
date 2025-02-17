package models

import (
	"time"
)

type Ingredients struct {
	BaseModel
	Name string `gorm:"type:varchar(256);not null" json:"name"`
}

type IngredientInBound struct {
	BaseModel
	IngredientId   *int         `gorm:"type:int(11)" json:"ingredientId"`
	Ingredient     *Ingredients `gorm:"foreignKey:IngredientId" json:"ingredient"`
	Supplier       string       `gorm:"type:varchar(256); DEFAULT ''" json:"supplier"`
	Specification  string       `gorm:"type:varchar(256)" json:"specification"`
	UnitPrice      float64      `gorm:"type:decimal(12,2)" json:"unitPrice"`
	TotalPrice     float64      `gorm:"type:decimal(12,2)" json:"totalPrice"`
	FinishPrice    float64      `gorm:"type:decimal(10,2)" json:"finishPrice"`
	PaymentHistory string       `gorm:"type:varchar(1024)" json:"paymentHistory"`
	Status         int          `gorm:"type:int(11);not null" json:"status"` // 1:未完成支付 2:已支付
	StockNum       float64      `gorm:"type:decimal(16,4)" json:"stockNum"`
	StockUnit      int          `gorm:"type:int(2)" json:"stockUnit"`
	StockUser      string       `gorm:"type:varchar(256)" json:"stockUser"`
	StockTime      time.Time    `gorm:"type:Time" json:"stockTime"`
}

type IngredientStock struct {
	BaseModel
	IngredientId *int         `gorm:"type:int(11)" json:"ingredientId"`
	Ingredient   *Ingredients `gorm:"foreignKey:IngredientId" json:"ingredient"`
	UnitPrice    float64      `gorm:"type:decimal(12,2)" json:"unitPrice"`
	StockNum     float64      `gorm:"type:decimal(16,4)" json:"stockNum"`
	StockUnit    int          `gorm:"type:int(2)" json:"stockUnit"`
}

type IngredientConsume struct {
	BaseModel
	IngredientId     *int         `gorm:"type:int(11)" json:"ingredientId"`
	Ingredient       *Ingredients `gorm:"foreignKey:IngredientId" json:"ingredient"`
	StockNum         float64      `gorm:"type:decimal(16,4)" json:"stockNum"`
	StockUnit        int          `gorm:"type:int(2)" json:"stockUnit"`
	OperationType    string       `gorm:"type:varchar(256)" json:"operationType"`
	OperationDetails string       `gorm:"type:varchar(256)" json:"operationDetails"`
	Cost             float64      `gorm:"type:decimal(12,2)" json:"cost"` // 成本
}

// 返回数据

// GetInBoundList 配料入库列表查询数据
type GetInBoundList struct {
	ID              int                 `json:"id"`
	Operator        string              `json:"operator"`
	Remark          string              `json:"remark"`
	CreatedAt       time.Time           `json:"createdAt"`
	UpdatedAt       time.Time           `json:"updatedAt"`
	IngredientId    *int                `json:"ingredientId"`
	Ingredient      *Ingredients        `json:"ingredient"`
	Supplier        string              `json:"supplier"`
	Specification   string              `json:"specification"`
	UnitPrice       float64             `json:"unitPrice"`
	TotalPrice      float64             `json:"totalPrice"`
	FinishPrice     float64             `json:"finishPrice"`
	UnFinishPrice   float64             `json:"unFinishPrice"`
	PaymentHistory  string              `json:"paymentHistory"`
	Status          int                 `json:"status"` // 1:未完成支付 2:已支付
	StockNum        float64             `json:"stockNum"`
	StockUnit       int                 `json:"stockUnit"`
	StockUser       string              `json:"stockUser"`
	StockTime       time.Time           `json:"stockTime"`
	FinishPriceList []map[string]string `json:"finishPriceList"`
}

// IngredientsUsage 出入库详情接口
type IngredientsUsage struct {
	IngredientId     int         `json:"ingredientId"`
	Ingredient       Ingredients `json:"ingredient"`
	StockNum         float64     `json:"stockNum"`
	StockUnit        int         `json:"stockUnit"`
	OperationType    string      `json:"operationType"`
	OperationDetails string      `json:"operationDetails"`
	Cost             float64     `json:"cost"` // 成本
	CreatedAt        time.Time   `json:"createdAt"`
}
