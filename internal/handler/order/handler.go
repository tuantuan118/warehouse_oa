package order

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"warehouse_oa/internal/handler"
	"warehouse_oa/internal/models"
	"warehouse_oa/internal/service"
	"warehouse_oa/utils"
)

type Order struct{}

var o Order

func InitOrderRouter(router *gin.RouterGroup) {
	orderRouter := router.Group("order")

	orderRouter.GET("list", o.list)
	orderRouter.GET("fields", o.fields)
	orderRouter.GET("export", o.export)
	orderRouter.GET("exportExecl", o.exportExecl)
	orderRouter.POST("add", o.add)
	orderRouter.POST("update", o.update)
	orderRouter.POST("finishOrder", o.finishOrder)
	orderRouter.POST("void", o.void)
	//orderRouter.POST("saveOutBound", o.saveOutBound)
}

func (*Order) list(c *gin.Context) {
	pn, pSize := utils.ParsePaginationParams(c)
	order := &models.Order{
		OrderNumber: c.DefaultQuery("orderNumber", ""),
		Salesman:    c.DefaultQuery("salesman", ""),
		Status:      utils.DefaultQueryInt(c, "status", 0),
	}
	customerStr := c.DefaultQuery("customerId", "")
	begTime := c.DefaultQuery("begTime", "")
	endTime := c.DefaultQuery("endTime", "")
	userId := c.GetInt("userId")

	data, err := service.GetOrderList(order, customerStr, begTime, endTime, pn, pSize, userId)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Order) add(c *gin.Context) {
	order := &models.Order{}
	if err := c.ShouldBindJSON(order); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	order.Operator = c.GetString("userName")
	data, err := service.SaveOrder(order)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Order) update(c *gin.Context) {
	order := &models.Order{}
	if err := c.ShouldBindJSON(order); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	order.Operator = c.GetString("userName")
	data, err := service.UpdateOrder(order)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

// 出库
func (*Order) finishOrder(c *gin.Context) {
	order := &models.Order{}
	if err := c.ShouldBindJSON(order); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	order.Operator = c.GetString("userName")
	err := service.FinishOrder(order.ID, order.Operator)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, nil)
}

//func (*Order) finishOrder(c *gin.Context) {
//	order := struct {
//		ID          int     `form:"id" json:"id" binding:"required"`
//		TotalPrice  float64 `json:"totalPrice"`
//		PaymentTime string  `json:"paymentTime"`
//		Operator    string  `json:"operator"`
//	}{}
//
//	if err := c.ShouldBindJSON(&order); err != nil {
//		// 如果解析失败，返回 400 错误和错误信息
//		handler.BadRequest(c, err.Error())
//		return
//	}
//
//	order.Operator = c.GetString("userName")
//	data, err := service.FinishOrder(order.ID, order.TotalPrice, order.PaymentTime, order.Operator)
//	if err != nil {
//		handler.InternalServerError(c, err)
//		return
//	}
//
//	handler.Success(c, data)
//}

func (*Order) void(c *gin.Context) {
	order := &models.Order{}
	if err := c.ShouldBindJSON(order); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	order.Operator = c.GetString("userName")
	err := service.VoidOrder(order.ID, order.Operator)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, nil)
}

func (*Order) export(c *gin.Context) {
	order := &models.Order{
		BaseModel: models.BaseModel{
			ID: utils.DefaultQueryInt(c, "id", 0),
		},
		OrderNumber: c.DefaultQuery("orderNumber", ""),
		CustomerId:  utils.DefaultQueryInt(c, "customerId", 0),
	}

	data, err := service.ExportOrder(order)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", "attachment; filename=file.pdf")

	c.Data(200, "application/pdf", data)
}

func (*Order) fields(c *gin.Context) {
	field := c.DefaultQuery("field", "")
	userId := c.GetInt("userId")

	data, err := service.GetOrderFieldList(field, userId)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Order) exportExecl(c *gin.Context) {
	pn, pSize := utils.ParsePaginationParams(c)
	order := &models.Order{
		OrderNumber: c.DefaultQuery("orderNumber", ""),
		Salesman:    c.DefaultQuery("salesman", ""),
		Status:      utils.DefaultQueryInt(c, "status", 0),
	}
	customerStr := c.DefaultQuery("customerId", "")
	begTime := c.DefaultQuery("begTime", "")
	endTime := c.DefaultQuery("endTime", "")
	userId := c.GetInt("userId")

	data, err := service.ExportOrderExecl(order, customerStr, begTime, endTime, pn, pSize, userId)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", `attachment; filename="配料入库.xlsx"`)
	c.Header("Content-Transfer-Encoding", "binary")

	// 将 Excel 文件写入到 HTTP 响应中
	if err = data.Write(c.Writer); err != nil {
		c.String(http.StatusInternalServerError, "文件生成失败")
		return
	}
}
