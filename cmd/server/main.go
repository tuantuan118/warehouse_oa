package main

import (
	"github.com/sirupsen/logrus"
	"warehouse_oa/internal/initialize"
	"warehouse_oa/internal/service"
)

func main() {
	if err := initialize.InitConfig(); err != nil {
		logrus.Panicf("init config err:%s", err.Error())
	}
	if err := initialize.InitDb(); err != nil {
		logrus.Panicf("init db err:%s", err.Error())
	}

	go service.Ticker()
	router := initialize.InitRouters()
	err := router.Run(":8090")
	if err != nil {
		logrus.Fatalln("Failed to start router", err.Error())
		return
	} // 监听并在 0.0.0.0:8080 上启动服务
}
