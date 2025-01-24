package ingredients

import (
	"github.com/gin-gonic/gin"
	"warehouse_oa/internal/handler"
	"warehouse_oa/internal/service"
	"warehouse_oa/utils"
)

type Inventory struct{}

var iv Inventory

func InitInventoryRouter(router *gin.RouterGroup) {
	inventoryRouter := router.Group("inventory")

	inventoryRouter.GET("list", iv.list)
	inventoryRouter.GET("fields", iv.fields)
}

func (*Inventory) list(c *gin.Context) {
	pn, pSize := utils.ParsePaginationParams(c)
	name := c.DefaultQuery("name", "")

	data, err := service.GetInventoryList(name, pn, pSize)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Inventory) fields(c *gin.Context) {
	field := c.DefaultQuery("field", "")
	data, err := service.GetInventoryFieldList(field)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}
