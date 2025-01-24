package finished

import (
	"github.com/gin-gonic/gin"
	"warehouse_oa/internal/handler"
	"warehouse_oa/internal/models"
	"warehouse_oa/internal/service"
	"warehouse_oa/utils"
)

type Manage struct{}

var m Manage

func InitManageRouter(router *gin.RouterGroup) {
	inBoundRouter := router.Group("manage")

	inBoundRouter.GET("list", m.list)
	inBoundRouter.GET("fields", m.fields)
	inBoundRouter.GET("getIngredients", m.getIngredientsByID)
	inBoundRouter.POST("add", m.add)
	inBoundRouter.POST("update", m.update)
	inBoundRouter.POST("delete", m.delete)
}

func (*Manage) list(c *gin.Context) {
	pn, pSize := utils.ParsePaginationParams(c)
	ids := c.DefaultQuery("id", "")
	name := c.DefaultQuery("name", "")

	data, err := service.GetFinishedManageList(ids, name, pn, pSize)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Manage) getIngredientsByID(c *gin.Context) {
	id := utils.DefaultQueryInt(c, "id", 0)

	data, err := service.GetFinishedManageIngredients(id)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Manage) add(c *gin.Context) {
	ingredients := &models.FinishedManage{}
	if err := c.ShouldBindJSON(ingredients); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	ingredients.Operator = c.GetString("userName")
	data, err := service.SaveFinishedManage(ingredients)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Manage) update(c *gin.Context) {
	ingredients := &models.FinishedManage{}
	if err := c.ShouldBindJSON(ingredients); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	ingredients.Operator = c.GetString("userName")
	data, err := service.UpdateFinishedManage(ingredients)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Manage) delete(c *gin.Context) {
	ingredients := &models.FinishedManage{}
	if err := c.ShouldBindJSON(ingredients); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	ingredients.Operator = c.GetString("userName")
	err := service.DelFinishedManage(ingredients.ID, ingredients.Operator)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, nil)
}

func (*Manage) fields(c *gin.Context) {
	field := c.DefaultQuery("field", "")
	data, err := service.GetFinishedManageFieldList(field)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}
