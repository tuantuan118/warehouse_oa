package ingredients

import (
	"github.com/gin-gonic/gin"
	"warehouse_oa/internal/handler"
	"warehouse_oa/internal/service"
	"warehouse_oa/utils"
)

type Stock struct{}

var iv Stock

func InitStockRouter(router *gin.RouterGroup) {
	stockRouter := router.Group("stock")

	stockRouter.GET("list", iv.list)
	stockRouter.GET("getStockName", iv.getStockName)
}

func (*Stock) list(c *gin.Context) {
	pn, pSize := utils.ParsePaginationParams(c)
	name := c.DefaultQuery("name", "")
	isPackage := utils.DefaultQueryInt(c, "isPackage", 0)

	data, err := service.GetStockList(name, pn, pSize, isPackage)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Stock) getStockName(c *gin.Context) {
	data, err := service.GetStockName()
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}
