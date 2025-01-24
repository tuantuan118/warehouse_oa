package models

type User struct {
	BaseModel
	Username string `gorm:"type:varchar(100);not null" json:"username"`
	Nickname string `gorm:"type:varchar(256)" json:"nickname"`
	Password string `gorm:"type:varchar(256);not null" json:"password,omitempty"`
	Roles    []Role `gorm:"many2many:user_role;" json:"roles"`
}

type UserRole struct {
	UserID int `gorm:"primaryKey;index"` // UserID 是联合主键并定义索引
	RoleID int `gorm:"primaryKey"`       // RoleID 也是联合主键
}
