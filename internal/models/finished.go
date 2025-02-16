package models

import "time"

type Finished struct {
	BaseModel
	Name     string             `gorm:"type:varchar(256);not null" json:"name"`
	Material []FinishedMaterial `gorm:"foreignKey:FinishedId;references:ID" json:"material"`
}

type FinishedMaterial struct {
	FinishedId   int          `gorm:"primaryKey;index" json:"finishedId"`
	IngredientId *int         `gorm:"type:int(11)" json:"ingredientId"`
	Ingredient   *Ingredients `gorm:"foreignKey:IngredientId" json:"ingredient"`
	StockUnit    int          `gorm:"type:int(2)" json:"stockUnit"`
	Quantity     float64      `gorm:"type:decimal(10,4);not null" json:"quantity"` // 用量
}

type FinishedProduction struct {
	BaseModel
	FinishedId         int        `gorm:"type:int(11)" json:"finishedId"`
	Finished           *Finished  `gorm:"foreignKey:FinishedId;" json:"finished"`
	Ratio              float64    `gorm:"type:decimal(10,2);not null" json:"ratio"`
	ExpectAmount       int        `gorm:"type:int(11);not null" json:"expectAmount"`
	ActualAmount       int        `gorm:"type:int(11);not null" json:"actualAmount"`
	Cost               float64    `gorm:"type:decimal(12,2)" json:"cost"` // 成本
	UnitPrice          float64    `gorm:"type:decimal(12,2)" json:"unitPrice"`
	Status             int        `gorm:"type:int(11);not null" json:"status"`
	EstimatedTime      *time.Time `gorm:"type:Time" json:"estimatedTime"`
	FinishTime         *time.Time `gorm:"type:Time" json:"finishTime"`
	ProductIngredients string     `gorm:"type:Text;not null" json:"productIngredients"`
}

type FinishedStock struct {
	BaseModel
	Name       string    `gorm:"type:varchar(256);not null" json:"name"`
	Amount     float64   `gorm:"type:decimal(10,2);not null" json:"amount"`
	FinishedId int       `gorm:"type:int(11)" json:"FinishedId"`
	Finished   *Finished `gorm:"foreignKey:FinishedId;" json:"finishedManage"`
	UnitPrice  float64   `gorm:"type:decimal(12,2)" json:"unitPrice"`
}

type FinishedConsume struct {
	BaseModel
	FinishedId       int       `gorm:"type:int(11)" json:"finishedId"`
	Finished         *Finished `gorm:"foreignKey:FinishedId;" json:"finished"`
	StockNum         float64   `gorm:"type:decimal(16,4)" json:"stockNum"`
	OperationType    string    `gorm:"type:varchar(256)" json:"operationType"`
	OperationDetails string    `gorm:"type:varchar(256)" json:"operationDetails"`
}
