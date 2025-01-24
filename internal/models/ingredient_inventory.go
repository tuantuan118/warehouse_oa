package models

type IngredientInventory struct {
	BaseModel
	IngredientID  *int         `gorm:"type:int(11)" json:"ingredientId"`
	Ingredient    *Ingredients `gorm:"foreignKey:IngredientID" json:"ingredient"`
	Name          string       `json:"name"`
	Specification string       `gorm:"type:varchar(256)" json:"specification"`
	Price         float64      `gorm:"type:decimal(12,2)" json:"price"`
	StockNum      float64      `gorm:"type:decimal(16,4)" json:"stockNum"`
	StockUnit     int          `gorm:"type:int(2)" json:"stockUnit"`
}
