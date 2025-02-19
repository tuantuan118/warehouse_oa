package finished

import (
	"github.com/gin-gonic/gin"
	"warehouse_oa/internal/handler"
	"warehouse_oa/internal/service"
	"warehouse_oa/utils"
)

type stock struct{}

var s stock

func InitFinishedStockRouter(router *gin.RouterGroup) {
	stockRouter := router.Group("stock")

	stockRouter.GET("list", s.list)
}

func (*stock) list(c *gin.Context) {
	pn, pSize := utils.ParsePaginationParams(c)
	ids := c.DefaultQuery("id", "")
	begTime := c.DefaultQuery("begTime", "")
	endTime := c.DefaultQuery("endTime", "")

	data, err := service.GetFinishedStockList(ids, begTime, endTime, pn, pSize)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}
