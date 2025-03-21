package ingredients

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"warehouse_oa/internal/handler"
	"warehouse_oa/internal/models"
	"warehouse_oa/internal/service"
	"warehouse_oa/utils"
)

type InBound struct{}

var ib InBound

func InitInBoundRouter(router *gin.RouterGroup) {
	inBoundRouter := router.Group("in_bound")

	inBoundRouter.GET("list", ib.list)
	inBoundRouter.GET("outList", ib.outList)
	inBoundRouter.GET("chart", ib.chart)
	inBoundRouter.GET("export", ib.export)
	inBoundRouter.GET("exportOut", ib.exportOut)
	inBoundRouter.GET("getSupplier", ib.getSupplier)
	inBoundRouter.POST("add", ib.add)
	inBoundRouter.POST("update", ib.update)
	inBoundRouter.POST("delete", ib.delete)
	inBoundRouter.POST("finishInBound", ib.finishInBound)
}

func (*InBound) list(c *gin.Context) {
	pn, pSize := utils.ParsePaginationParams(c)
	name := c.DefaultQuery("name", "")
	supplier := c.DefaultQuery("supplier", "")
	stockUnit := c.DefaultQuery("stockUnit", "")
	begTime := c.DefaultQuery("begTime", "")
	endTime := c.DefaultQuery("endTime", "")

	data, err := service.GetInBoundList(name, stockUnit, supplier,
		begTime, endTime, pn, pSize)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*InBound) add(c *gin.Context) {
	ingredients := &models.IngredientInBound{}
	if err := c.ShouldBindJSON(ingredients); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	ingredients.Operator = c.GetString("userName")
	data, err := service.SaveInBound(ingredients)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*InBound) update(c *gin.Context) {
	ingredients := &models.IngredientInBound{}
	if err := c.ShouldBindJSON(ingredients); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	ingredients.Operator = c.GetString("userName")
	data, err := service.UpdateInBound(ingredients)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*InBound) delete(c *gin.Context) {
	ingredients := &models.IngredientInBound{}
	if err := c.ShouldBindJSON(ingredients); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	ingredients.Operator = c.GetString("userName")
	err := service.DelInBound(ingredients.ID, ingredients.Operator)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, nil)
}

func (*InBound) finishInBound(c *gin.Context) {
	inBound := struct {
		ID          int     `form:"id" json:"id" binding:"required"`
		TotalPrice  float64 `json:"totalPrice"`
		PaymentTime string  `json:"paymentTime"`
		Operator    string  `json:"operator"`
	}{}

	if err := c.ShouldBindJSON(&inBound); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	inBound.Operator = c.GetString("userName")
	data, err := service.FinishInBound(inBound.ID, inBound.TotalPrice, inBound.PaymentTime, inBound.Operator)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*InBound) export(c *gin.Context) {
	name := c.DefaultQuery("name", "")
	stockUser := c.DefaultQuery("stockUser", "")
	begTime := c.DefaultQuery("begTime", "")
	endTime := c.DefaultQuery("endTime", "")

	data, err := service.ExportIngredients(name, stockUser, begTime, endTime)
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

func (*InBound) exportOut(c *gin.Context) {
	ids := c.DefaultQuery("ids", "")
	stockUnit := c.DefaultQuery("stockUnit", "")
	begTime := c.DefaultQuery("begTime", "")
	endTime := c.DefaultQuery("endTime", "")

	data, err := service.ExportConsume(ids, stockUnit, begTime, endTime)
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

func (*InBound) outList(c *gin.Context) {
	pn, pSize := utils.ParsePaginationParams(c)
	ids := c.DefaultQuery("ids", "")
	stockUnit := c.DefaultQuery("stockUnit", "")
	begTime := c.DefaultQuery("begTime", "")
	endTime := c.DefaultQuery("endTime", "")
	inOrOut := utils.DefaultQueryInt(c, "inOrOut", 0)

	data, err := service.GetConsumeList(ids, stockUnit, begTime, endTime, inOrOut, pn, pSize)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*InBound) chart(c *gin.Context) {
	ids := c.DefaultQuery("ids", "")
	stockUnit := c.DefaultQuery("stockUnit", "")
	begTime := c.DefaultQuery("begTime", "")
	endTime := c.DefaultQuery("endTime", "")

	data, err := service.GetConsumeChart(ids, stockUnit, begTime, endTime)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*InBound) getSupplier(c *gin.Context) {
	data, err := service.GetSupplier()
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}
