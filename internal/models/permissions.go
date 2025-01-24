package models

type Permission struct {
	BaseModel
	Name     string      `gorm:"type:varchar(100);not null" json:"name"`
	NameEn   string      `gorm:"type:varchar(100);not null" json:"nameEn"`
	Url      string      `gorm:"type:varchar(256);not null" json:"url"`
	Coding   string      `gorm:"type:varchar(100)" json:"coding"`
	Order    int         `gorm:"type:int(11)" json:"order"`
	Type     int         `gorm:"type:int(11)" json:"type"`              // 1、菜单 2、页面 3、按钮 4、字段
	Enabled  bool        `gorm:"type:bool;default:true" json:"enabled"` // true 表示启用，false 表示禁用
	ParentID *int        `gorm:"type:int(11)" json:"parentId"`          // 父权限的ID，允许为 NULL
	Parent   *Permission `gorm:"foreignKey:ParentID" json:"-"`          // 自关联，通过 ParentID 关联父权限
}
