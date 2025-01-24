package ingredients

import (
	"github.com/gin-gonic/gin"
	"warehouse_oa/internal/handler"
	"warehouse_oa/internal/models"
	"warehouse_oa/internal/service"
	"warehouse_oa/utils"
)

type Ingredients struct{}

var i Ingredients

func InitIngredientsRouter(router *gin.RouterGroup) {
	ingredientsRouter := router.Group("ingredients")

	ingredientsRouter.GET("list", i.list)
	ingredientsRouter.GET("fields", i.fields)
	ingredientsRouter.POST("add", i.add)
	ingredientsRouter.POST("update", i.update)
	ingredientsRouter.POST("delete", i.delete)
}

func (*Ingredients) list(c *gin.Context) {
	pn, pSize := utils.ParsePaginationParams(c)
	ingredients := &models.Ingredients{
		Name: c.DefaultQuery("name", ""),
	}
	data, err := service.GetIngredientsList(ingredients, pn, pSize)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Ingredients) add(c *gin.Context) {
	ingredients := &models.Ingredients{}
	if err := c.ShouldBindJSON(ingredients); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	ingredients.Operator = c.GetString("userName")
	data, err := service.SaveIngredients(ingredients)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Ingredients) update(c *gin.Context) {
	ingredients := &models.Ingredients{}
	if err := c.ShouldBindJSON(ingredients); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	ingredients.Operator = c.GetString("userName")
	data, err := service.UpdateIngredients(ingredients)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Ingredients) delete(c *gin.Context) {
	ingredients := &models.Ingredients{}
	if err := c.ShouldBindJSON(ingredients); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	ingredients.Operator = c.GetString("userName")
	err := service.DelIngredients(ingredients.ID, ingredients.Operator)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, nil)
}

func (*Ingredients) fields(c *gin.Context) {
	field := c.DefaultQuery("field", "")
	data, err := service.GetIngredientsFieldList(field)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}
