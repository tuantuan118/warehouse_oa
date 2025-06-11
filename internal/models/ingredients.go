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
	IngredientId *int               `gorm:"type:int(11)" json:"ingredientId"`
	Ingredient   *Ingredients       `gorm:"foreignKey:IngredientId" json:"ingredient"`
	InBoundId    *int               `gorm:"type:int(11)" json:"inBoundId"`
	InBound      *IngredientInBound `gorm:"foreignKey:InBoundId" json:"inBound"`
	StockNum     float64            `gorm:"type:decimal(16,4)" json:"stockNum"`
	StockUnit    int                `gorm:"type:int(2)" json:"stockUnit"`
}

type IngredientConsume struct {
	BaseModel
	FinishedId       *int                `gorm:"type:int(11)" json:"finishedId"` // 关联成品ID
	Finish           *Finished           `gorm:"foreignKey:FinishedId" json:"finish"`
	IngredientId     *int                `gorm:"type:int(11)" json:"ingredientId"` // 关联配料ID
	Ingredient       *Ingredients        `gorm:"foreignKey:IngredientId" json:"ingredient"`
	ProductionId     *int                `gorm:"type:int(11)" json:"productionId"` // 关联报功ID
	Production       *FinishedProduction `gorm:"foreignKey:ProductionId" json:"production"`
	InBoundId        *int                `gorm:"type:int(11)" json:"inBoundId"` // 关联入库ID
	InBound          *IngredientInBound  `gorm:"foreignKey:InBoundId" json:"inBound"`
	OrderId          *int                `gorm:"type:int(11)" json:"OrderId"` // 订单ID
	StockNum         float64             `gorm:"type:decimal(16,4)" json:"stockNum"`
	StockUnit        int                 `gorm:"type:int(2)" json:"stockUnit"`
	OperationType    bool                `gorm:"type:bool" json:"operationType"` // true表示启用，false表示禁用
	OperationDetails string              `gorm:"type:varchar(256)" json:"operationDetails"`

	// 返回参数
	Cost float64 `gorm:"-" json:"cost"`
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
	OperationType    bool        `json:"operationType"` // true 表示启用，false 表示禁用
	OperationDetails string      `json:"operationDetails"`
	Cost             float64     `json:"cost"` // 成本
	CreatedAt        time.Time   `json:"createdAt"`
}

type FinishInBound struct {
	ID          int     `form:"id" json:"id" binding:"required"`
	TotalPrice  float64 `json:"totalPrice"`
	PaymentTime string  `json:"paymentTime"`
	Operator    string  `json:"operator"`
}
