package ingredients

import "github.com/gin-gonic/gin"

func InitIngredientRouter(router *gin.RouterGroup) {
	ingredientRouter := router.Group("ingredient")

	InitInBoundRouter(ingredientRouter)
	InitIngredientsRouter(ingredientRouter)
	InitStockRouter(ingredientRouter)
}
