package models

type FinishedManage struct {
	BaseModel
	Name     string             `gorm:"type:varchar(256);not null" json:"name"`
	Material []FinishedMaterial `gorm:"foreignKey:FinishedManageID;references:ID" json:"material"`
}

type FinishedMaterial struct {
	FinishedManageID    int                  `gorm:"primaryKey;index" json:"finished_manage_id"`
	IngredientID        int                  `gorm:"primaryKey;" json:"ingredient_id"`
	IngredientInventory *IngredientInventory `gorm:"foreignKey:IngredientID" json:"ingredient_inventory"`
	Quantity            float64              `gorm:"type:decimal(10,4);not null" json:"quantity"` // 用量
}
