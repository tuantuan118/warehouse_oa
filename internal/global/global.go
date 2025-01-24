package global

import (
	"gorm.io/gorm"
)

var (
	Db           *gorm.DB
	ServerConfig = &ServerConfigInfo{}
)

type MysqlConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	DbName   string `json:"db_name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type JWTConfig struct {
	SigningKey string `json:"signing_key"`
}

type ServerConfigInfo struct {
	JWTInfo   JWTConfig   `json:"jwt"`
	MysqlInfo MysqlConfig `json:"mysql"`
}
