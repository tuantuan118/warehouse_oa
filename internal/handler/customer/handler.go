package customer

import (
	"github.com/gin-gonic/gin"
	"warehouse_oa/internal/handler"
	"warehouse_oa/internal/models"
	"warehouse_oa/internal/service"
	"warehouse_oa/utils"
)

type Customer struct{}

var c Customer

func InitCustomerRouter(router *gin.RouterGroup) {
	customerRouter := router.Group("customer")

	customerRouter.GET("list", c.list)
	customerRouter.GET("fields", c.fields)
	customerRouter.POST("add", c.add)
	customerRouter.POST("update", c.update)
	customerRouter.POST("delete", c.delete)
}

func (*Customer) list(c *gin.Context) {
	pn, pSize := utils.ParsePaginationParams(c)
	customer := &models.Customer{
		Name: c.DefaultQuery("name", ""),
	}
	data, err := service.GetCustomerList(customer, pn, pSize)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Customer) add(c *gin.Context) {
	customer := &models.Customer{}
	if err := c.ShouldBindJSON(customer); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	customer.Operator = c.GetString("userName")
	data, err := service.SaveCustomer(customer)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Customer) update(c *gin.Context) {
	customer := &models.Customer{}
	if err := c.ShouldBindJSON(customer); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	customer.Operator = c.GetString("userName")
	data, err := service.UpdateCustomer(customer)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Customer) delete(c *gin.Context) {
	customer := &models.Customer{}
	if err := c.ShouldBindJSON(customer); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	customer.Operator = c.GetString("userName")
	err := service.DelCustomer(customer.ID, customer.Operator)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, nil)
}

func (*Customer) fields(c *gin.Context) {
	field := c.DefaultQuery("field", "")
	data, err := service.GetCustomerFieldList(field)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}
