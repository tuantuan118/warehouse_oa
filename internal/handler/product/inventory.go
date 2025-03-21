package product

import (
	"github.com/gin-gonic/gin"
	"warehouse_oa/internal/handler"
	"warehouse_oa/internal/models"
	"warehouse_oa/internal/service"
	"warehouse_oa/utils"
)

type Inventory struct{}

var i Inventory

func InitInventoryRouter(router *gin.RouterGroup) {
	inventoryRouter := router.Group("inventory")

	inventoryRouter.GET("list", i.list)
	inventoryRouter.POST("add", i.add)
	inventoryRouter.POST("update", i.update)
}

func (*Inventory) list(c *gin.Context) {
	pn, pSize := utils.ParsePaginationParams(c)
	inventory := &models.ProductInventory{
		BaseModel: models.BaseModel{
			ID: utils.DefaultQueryInt(c, "id", 0),
		},
		ProductId: utils.DefaultQueryInt(c, "productId", 0),
	}
	data, err := service.GetProductInventoryList(inventory, pn, pSize)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Inventory) add(c *gin.Context) {
	inventory := &models.ProductInventory{}
	if err := c.ShouldBindJSON(inventory); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	inventory.Operator = c.GetString("userName")
	err := service.SaveProductInventory(inventory)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, nil)
}

func (*Inventory) update(c *gin.Context) {
	inventory := &models.ProductInventory{}
	if err := c.ShouldBindJSON(inventory); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	inventory.Operator = c.GetString("userName")
	err := service.UpdateProductInventory(inventory)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, nil)
}
