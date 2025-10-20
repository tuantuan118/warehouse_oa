package initialize

import "warehouse_oa/internal/global"

func InitConfig() error {
	global.ServerConfig.MysqlInfo.Host = "8.138.155.131"
	global.ServerConfig.MysqlInfo.Port = 3306
	global.ServerConfig.MysqlInfo.Username = "ware"
	global.ServerConfig.MysqlInfo.Password = "ware123456"
	global.ServerConfig.MysqlInfo.DbName = "warehouse_oa"

	return nil
}
