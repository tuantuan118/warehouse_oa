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
	finishedRouter := router.Group("production")

	finishedRouter.GET("list", p.list)
	finishedRouter.GET("outList", p.outList)
	finishedRouter.GET("fields", p.fields)
	finishedRouter.POST("add", p.add)
	finishedRouter.POST("update", p.update)
	finishedRouter.POST("delete", p.delete)
	finishedRouter.POST("void", p.void)
	finishedRouter.POST("finish", p.finish)
}

func (*Production) list(c *gin.Context) {
	pn, pSize := utils.ParsePaginationParams(c)
	finished := &models.FinishedProduction{
		Name:   c.DefaultQuery("name", ""),
		Status: utils.DefaultQueryInt(c, "status", -1),
	}
	begTime := c.DefaultQuery("begTime", "")
	endTime := c.DefaultQuery("endTime", "")

	data, err := service.GetFinishedList(finished,
		begTime, endTime,
		pn, pSize, true)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Production) add(c *gin.Context) {
	finished := &models.FinishedProduction{}
	if err := c.ShouldBindJSON(finished); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	finished.Operator = c.GetString("userName")
	data, err := service.SaveFinished(finished)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Production) update(c *gin.Context) {
	finished := &models.FinishedProduction{}
	if err := c.ShouldBindJSON(finished); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	finished.Operator = c.GetString("userName")
	data, err := service.UpdateFinished(finished)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Production) delete(c *gin.Context) {
	finished := &models.FinishedProduction{}
	if err := c.ShouldBindJSON(finished); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	finished.Operator = c.GetString("userName")
	err := service.DelFinished(finished.ID, finished.Operator)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, nil)
}

func (*Production) void(c *gin.Context) {
	finished := &models.FinishedProduction{}
	if err := c.ShouldBindJSON(finished); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	finished.Operator = c.GetString("userName")
	err := service.VoidFinished(finished.ID, finished.Operator)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, nil)
}

func (*Production) finish(c *gin.Context) {
	finished := &models.FinishedProduction{}
	if err := c.ShouldBindJSON(finished); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	finished.Operator = c.GetString("userName")
	err := service.FinishFinished(finished.ID, finished.ActualAmount, finished.Operator)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, nil)
}

func (*Production) fields(c *gin.Context) {
	field := c.DefaultQuery("field", "")
	data, err := service.GetProductionFieldList(field)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Production) outList(c *gin.Context) {
	pn, pSize := utils.ParsePaginationParams(c)
	finished := &models.FinishedProduction{
		BaseModel: models.BaseModel{
			ID: utils.DefaultQueryInt(c, "id", 0),
		},
		Name:   c.DefaultQuery("name", ""),
		Status: utils.DefaultQueryInt(c, "status", -1),
	}
	begTime := c.DefaultQuery("begTime", "")
	endTime := c.DefaultQuery("endTime", "")

	data, err := service.GetOutFinishedList(finished,
		begTime, endTime,
		pn, pSize)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}
