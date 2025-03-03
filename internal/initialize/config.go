package initialize

import "warehouse_oa/internal/global"

func InitConfig() error {
	global.ServerConfig.MysqlInfo.Host = "0.0.0.0"
	global.ServerConfig.MysqlInfo.Port = 3306
	global.ServerConfig.MysqlInfo.Username = "root"
	global.ServerConfig.MysqlInfo.Password = "123456"
	global.ServerConfig.MysqlInfo.DbName = "warehouse_oa"

	return nil
}
