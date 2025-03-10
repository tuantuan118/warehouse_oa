package v1

import (
	"github.com/gin-gonic/gin"
	"warehouse_oa/internal/handler"
	"warehouse_oa/internal/service"
)

func InitV1Router(router *gin.RouterGroup) {
	router.GET("queryUpdate", queryUpdate)
}

func queryUpdate(c *gin.Context) {
	update := c.DefaultQuery("update", "")
	updateTime := c.DefaultQuery("updateTime", "")
	data, err := service.GetUpdate(update, updateTime)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}
