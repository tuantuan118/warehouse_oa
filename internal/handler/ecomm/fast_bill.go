package ecomm

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"warehouse_oa/internal/handler"
	"warehouse_oa/internal/models"
	"warehouse_oa/internal/service"
	"warehouse_oa/utils"
)

type FastBillHandler struct{}

var f FastBillHandler

func InitFastBillRouter(router *gin.RouterGroup) {
	fastBillRouter := router.Group("fastBill")

	fastBillRouter.GET("list", f.list)
	fastBillRouter.GET("fields", f.fields)
	fastBillRouter.POST("add", f.add)
	fastBillRouter.POST("update", f.update)
	fastBillRouter.POST("delete", f.delete)
	fastBillRouter.POST("upload", f.upload)
}

func (*FastBillHandler) list(c *gin.Context) {
	pn, pSize := utils.ParsePaginationParams(c)
	fastBill := &models.FastBill{
		OrderNumber: c.DefaultQuery("orderNumber", ""),
	}
	data, err := service.GetFastBillList(fastBill, pn, pSize)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*FastBillHandler) add(c *gin.Context) {
	fastBill := &models.FastBill{}
	if err := c.ShouldBindJSON(fastBill); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	fastBill.Operator = c.GetString("userName")
	data, err := service.SaveFastBill(fastBill)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*FastBillHandler) update(c *gin.Context) {
	fastBill := &models.FastBill{}
	if err := c.ShouldBindJSON(fastBill); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	fastBill.Operator = c.GetString("userName")
	data, err := service.UpdateFastBill(fastBill)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*FastBillHandler) delete(c *gin.Context) {
	fastBill := &models.FastBill{}
	if err := c.ShouldBindJSON(fastBill); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	err := service.DelFastBill(fastBill.ID)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, nil)
}

func (*FastBillHandler) fields(c *gin.Context) {
	field := c.DefaultQuery("field", "")
	data, err := service.GetFastBillFieldList(field)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*FastBillHandler) upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, "文件上传失败: %v", err)
		return
	}

	username := c.GetString("userName")
	err = service.UploadFastBill(file, username)
	if err != nil {
		c.String(http.StatusBadRequest, "文件读取失败: %v", err)
		return
	}

	handler.Success(c, nil)
}
