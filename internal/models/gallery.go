package models

type Gallery struct {
	BaseModel
	Name string `gorm:"type:varchar(256)" json:"name"`
	Url  string `gorm:"type:varchar(500)" json:"url"`
}
