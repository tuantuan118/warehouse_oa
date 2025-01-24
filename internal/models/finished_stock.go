package models

type FinishedStock struct {
	BaseModel
	Name               string          `gorm:"type:varchar(256);not null" json:"name"`
	Amount             float64         `gorm:"type:decimal(10,2);not null" json:"amount"`
	ProductIngredients string          `gorm:"type:Text;not null" json:"productIngredients"`
	FinishedManageId   int             `gorm:"type:int(11)" json:"finishedManageId"`
	FinishedManage     *FinishedManage `gorm:"foreignKey:FinishedManageId;" json:"finishedManage"`
}
