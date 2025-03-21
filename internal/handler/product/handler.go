package product

import "github.com/gin-gonic/gin"

func InitAllProductRouter(router *gin.RouterGroup) {
	productRouter := router.Group("product")

	InitProductRouter(productRouter)
	InitInventoryRouter(productRouter)
}
