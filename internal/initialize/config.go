package initialize

import "warehouse_oa/internal/global"

func InitConfig() error {
	global.ServerConfig.MysqlInfo.Host = "127.0.0。1"
	global.ServerConfig.MysqlInfo.Port = 3306
	global.ServerConfig.MysqlInfo.Username = "root"
	global.ServerConfig.MysqlInfo.Password = "123456"
	global.ServerConfig.MysqlInfo.DbName = "warehouse_db"

	return nil
}
