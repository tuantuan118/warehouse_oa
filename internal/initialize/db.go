package initialize

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"time"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

func InitDb() error {
	mysqlInfo := global.ServerConfig.MysqlInfo
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		mysqlInfo.Username,
		mysqlInfo.Password,
		mysqlInfo.Host,
		mysqlInfo.Port,
		mysqlInfo.DbName,
	)

	var ormLogger logger.Interface
	if gin.Mode() == "debug" {
		ormLogger = logger.Default.LogMode(logger.Info)
	} else {
		ormLogger = logger.Default
	}

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:               dsn,
		DefaultStringSize: 256,
	}), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "tb_",
			SingularTable: true,
		},
		Logger: ormLogger,
	})
	if err != nil {
		logrus.Error("mysql connection failed, err: ", err.Error())
		return err
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(time.Second * 30)
	global.Db = db

	migration()
	return nil
}

func migration() {
	err := global.Db.Set("gorm:table_options", "charset=utf8mb4").AutoMigrate(
		&models.Customer{},
		&models.IngredientInBound{},
		&models.IngredientInventory{},
		&models.Ingredients{},
		&models.Order{},
		&models.Permission{},
		&models.Finished{},
		&models.FinishedManage{},
		&models.FinishedStock{},
		&models.FinishedMaterial{},
		&models.Role{},
		&models.User{},
		&models.ECommBill{},
		&models.ECommCustomers{},
		&models.FastBill{},
		&models.Gallery{},
		&models.Product{},
		&models.AddIngredient{},
	)
	if err != nil {
		logrus.Error("migration err: ", err.Error())
	}
}
