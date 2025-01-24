package ecomm

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"warehouse_oa/internal/handler"
	"warehouse_oa/internal/models"
	"warehouse_oa/internal/service"
	"warehouse_oa/utils"
)

type BillHandler struct{}

var e BillHandler

func InitECommBillRouter(router *gin.RouterGroup) {
	eCommBillRouter := router.Group("bill")

	eCommBillRouter.GET("list", e.list)
	eCommBillRouter.GET("fields", e.fields)
	eCommBillRouter.POST("add", e.add)
	eCommBillRouter.POST("update", e.update)
	eCommBillRouter.POST("delete", e.delete)
	eCommBillRouter.POST("upload", e.upload)
}

func (*BillHandler) list(c *gin.Context) {
	pn, pSize := utils.ParsePaginationParams(c)
	eCommBill := &models.ECommBill{
		Name: c.DefaultQuery("name", ""),
	}
	data, err := service.GetECommBillList(eCommBill, pn, pSize)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*BillHandler) add(c *gin.Context) {
	eCommBill := &models.ECommBill{}
	if err := c.ShouldBindJSON(eCommBill); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	eCommBill.Operator = c.GetString("userName")
	data, err := service.SaveECommBill(eCommBill)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*BillHandler) update(c *gin.Context) {
	eCommBill := &models.ECommBill{}
	if err := c.ShouldBindJSON(eCommBill); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	eCommBill.Operator = c.GetString("userName")
	data, err := service.UpdateECommBill(eCommBill)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*BillHandler) delete(c *gin.Context) {
	eCommBill := &models.ECommBill{}
	if err := c.ShouldBindJSON(eCommBill); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	eCommBill.Operator = c.GetString("userName")
	err := service.DelECommBill(eCommBill.ID, eCommBill.Operator)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, nil)
}

func (*BillHandler) fields(c *gin.Context) {
	field := c.DefaultQuery("field", "")
	data, err := service.GetECommBillFieldList(field)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*BillHandler) upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, "文件上传失败: %v", err)
		return
	}

	username := c.GetString("userName")
	err = service.UploadECommBill(file, username)
	if err != nil {
		c.String(http.StatusBadRequest, "文件读取失败: %v", err)
		return
	}

	handler.Success(c, nil)
}
