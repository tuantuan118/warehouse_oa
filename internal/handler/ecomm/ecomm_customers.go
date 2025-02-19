package ecomm

import (
	"github.com/gin-gonic/gin"
	"warehouse_oa/internal/handler"
	"warehouse_oa/internal/models"
	"warehouse_oa/internal/service"
	"warehouse_oa/utils"
)

type CustomersHandler struct{}

var c CustomersHandler

func InitECommCustomersRouter(router *gin.RouterGroup) {
	eCommCustomersRouter := router.Group("customers")

	eCommCustomersRouter.GET("list", c.list)
	eCommCustomersRouter.GET("fields", c.fields)
	eCommCustomersRouter.POST("add", c.add)
	eCommCustomersRouter.POST("update", c.update)
	eCommCustomersRouter.POST("delete", c.delete)
}

func (*CustomersHandler) list(c *gin.Context) {
	pn, pSize := utils.ParsePaginationParams(c)
	eCommCustomers := &models.ECommCustomers{
		Name: c.DefaultQuery("name", ""),
	}
	data, err := service.GetECommCustomersList(eCommCustomers, pn, pSize)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*CustomersHandler) add(c *gin.Context) {
	eCommCustomers := &models.ECommCustomers{}
	if err := c.ShouldBindJSON(eCommCustomers); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	eCommCustomers.Operator = c.GetString("userName")
	data, err := service.SaveECommCustomers(eCommCustomers)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*CustomersHandler) update(c *gin.Context) {
	eCommCustomers := &models.ECommCustomers{}
	if err := c.ShouldBindJSON(eCommCustomers); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	eCommCustomers.Operator = c.GetString("userName")
	data, err := service.UpdateECommCustomers(eCommCustomers)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*CustomersHandler) delete(c *gin.Context) {
	eCommCustomers := &models.ECommCustomers{}
	if err := c.ShouldBindJSON(eCommCustomers); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	err := service.DelECommCustomers(eCommCustomers.ID)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, nil)
}

func (*CustomersHandler) fields(c *gin.Context) {
	field := c.DefaultQuery("field", "")
	data, err := service.GetECommCustomersFieldList(field)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}
