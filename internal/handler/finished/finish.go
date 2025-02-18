package finished

import (
	"github.com/gin-gonic/gin"
	"warehouse_oa/internal/handler"
	"warehouse_oa/internal/models"
	"warehouse_oa/internal/service"
	"warehouse_oa/utils"
)

type Finish struct{}

var f Finish

func InitFinishRouter(router *gin.RouterGroup) {
	inBoundRouter := router.Group("finished")

	inBoundRouter.GET("list", f.list)
	inBoundRouter.GET("fields", f.fields)
	inBoundRouter.GET("getIngredients", f.getIngredientsByID)
	inBoundRouter.POST("add", f.add)
	inBoundRouter.POST("update", f.update)
	inBoundRouter.POST("delete", f.delete)
}

func (*Finish) list(c *gin.Context) {
	pn, pSize := utils.ParsePaginationParams(c)
	ids := c.DefaultQuery("id", "")
	name := c.DefaultQuery("name", "")

	data, err := service.GetFinishedList(ids, name, pn, pSize)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Finish) getIngredientsByID(c *gin.Context) {
	id := utils.DefaultQueryInt(c, "id", 0)

	data, err := service.GetFinishedIngredients(id)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Finish) add(c *gin.Context) {
	ingredients := &models.Finished{}
	if err := c.ShouldBindJSON(ingredients); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	ingredients.Operator = c.GetString("userName")
	data, err := service.SaveFinished(ingredients)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Finish) update(c *gin.Context) {
	ingredients := &models.Finished{}
	if err := c.ShouldBindJSON(ingredients); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	ingredients.Operator = c.GetString("userName")
	data, err := service.UpdateFinished(ingredients)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Finish) delete(c *gin.Context) {
	ingredients := &models.Finished{}
	if err := c.ShouldBindJSON(ingredients); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	err := service.DelFinished(ingredients.ID)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, nil)
}

func (*Finish) fields(c *gin.Context) {
	field := c.DefaultQuery("field", "")
	data, err := service.GetFinishedFieldList(field)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}
