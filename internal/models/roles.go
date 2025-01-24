package models

type Role struct {
	BaseModel
	Name        string       `gorm:"type:varchar(100);not null" json:"name"`
	NameEn      string       `gorm:"type:varchar(100);not null" json:"nameEn"`
	Enabled     bool         `gorm:"type:bool;default:true" json:"enabled"` // true 表示启用，false 表示禁用
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions"`
}

type RolePermission struct {
	RoleID       int `gorm:"primaryKey;index"` // RoleID 是联合主键并定义索引
	PermissionID int `gorm:"primaryKey"`       // PermissionID 也是联合主键
}
