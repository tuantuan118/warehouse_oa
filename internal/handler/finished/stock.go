package finished

import (
	"github.com/gin-gonic/gin"
	"warehouse_oa/internal/handler"
	"warehouse_oa/internal/models"
	"warehouse_oa/internal/service"
	"warehouse_oa/utils"
)

type stock struct{}

var s stock

func InitFinishedStockRouter(router *gin.RouterGroup) {
	stockRouter := router.Group("stock")

	stockRouter.GET("list", s.list)
	stockRouter.GET("fields", s.fields)
}

func (*stock) list(c *gin.Context) {
	pn, pSize := utils.ParsePaginationParams(c)
	finishedStock := &models.FinishedStock{
		Name: c.DefaultQuery("name", ""),
	}
	begTime := c.DefaultQuery("begTime", "")
	endTime := c.DefaultQuery("endTime", "")

	data, err := service.GetFinishedStockList(finishedStock, begTime, endTime, pn, pSize)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*stock) fields(c *gin.Context) {
	field := c.DefaultQuery("field", "")
	data, err := service.GetFinishedStockFieldList(field)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}
