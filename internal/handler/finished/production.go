package finished

import (
	"github.com/gin-gonic/gin"
	"warehouse_oa/internal/handler"
	"warehouse_oa/internal/models"
	"warehouse_oa/internal/service"
	"warehouse_oa/utils"
)

type Production struct{}

var p Production

func InitProductionRouter(router *gin.RouterGroup) {
	productionRouter := router.Group("production")

	productionRouter.GET("list", p.list)
	productionRouter.GET("outList", p.outList)
	productionRouter.POST("add", p.add)
	productionRouter.POST("update", p.update)
	productionRouter.POST("delete", p.delete)
	productionRouter.POST("void", p.void)
	productionRouter.POST("finish", p.finish)
}

// list 成品报工列表
func (*Production) list(c *gin.Context) {
	pn, pSize := utils.ParsePaginationParams(c)
	production := &models.FinishedProduction{
		FinishedId: utils.DefaultQueryInt(c, "finishedId", 0),
		Status:     utils.DefaultQueryInt(c, "status", -1),
	}
	begTime := c.DefaultQuery("begTime", "")
	endTime := c.DefaultQuery("endTime", "")

	data, err := service.GetProductionList(production,
		begTime, endTime, pn, pSize)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Production) add(c *gin.Context) {
	production := &models.FinishedProduction{}
	if err := c.ShouldBindJSON(production); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	production.Operator = c.GetString("userName")
	data, err := service.SaveProduction(production)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Production) update(c *gin.Context) {
	production := &models.FinishedProduction{}
	if err := c.ShouldBindJSON(production); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	production.Operator = c.GetString("userName")
	data, err := service.UpdateFinished(production)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Production) delete(c *gin.Context) {
	production := &models.FinishedProduction{}
	if err := c.ShouldBindJSON(production); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	production.Operator = c.GetString("userName")
	err := service.DelFinished(production.ID, production.Operator)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, nil)
}

func (*Production) void(c *gin.Context) {
	production := &models.FinishedProduction{}
	if err := c.ShouldBindJSON(production); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	production.Operator = c.GetString("userName")
	err := service.VoidFinished(production.ID, production.Operator)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, nil)
}

func (*Production) finish(c *gin.Context) {
	production := &models.FinishedProduction{}
	if err := c.ShouldBindJSON(production); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	production.Operator = c.GetString("userName")
	err := service.FinishFinished(production.ID, production.ActualAmount, production.Operator)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, nil)
}

// outList 成品出入库接口
func (*Production) outList(c *gin.Context) {
	pn, pSize := utils.ParsePaginationParams(c)
	production := &models.FinishedProduction{
		BaseModel: models.BaseModel{
			ID: utils.DefaultQueryInt(c, "id", 0),
		},
		FinishedId: utils.DefaultQueryInt(c, "finishedId", 0),
		Status:     utils.DefaultQueryInt(c, "status", -1),
	}
	begTime := c.DefaultQuery("begTime", "")
	endTime := c.DefaultQuery("endTime", "")

	data, err := service.GetFinishedConsumeList(production,
		begTime, endTime,
		pn, pSize)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}
