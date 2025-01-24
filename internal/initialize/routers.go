package initialize

import (
	"github.com/gin-gonic/gin"
	"warehouse_oa/internal/handler/customer"
	"warehouse_oa/internal/handler/ecomm"
	"warehouse_oa/internal/handler/finished"
	"warehouse_oa/internal/handler/gallery"
	"warehouse_oa/internal/handler/ingredients"
	"warehouse_oa/internal/handler/order"
	"warehouse_oa/internal/handler/product"
	"warehouse_oa/internal/handler/user"
	"warehouse_oa/internal/middlewares"
)

func InitRouters() *gin.Engine {
	Router := gin.Default()
	Router.Static("/images", "./cos/images")
	Router.Use(middlewares.Cors())

	apiGroup := Router.Group("/api/v1")
	user.InitLoginRouter(apiGroup)

	group := apiGroup
	group.Use(middlewares.JWTAuth())
	{
		user.InitUserRouter(group)
		user.InitRoleRouter(group)

		customer.InitCustomerRouter(group)
		ingredients.InitIngredientRouter(group)
		finished.InitFinishedAllRouter(group)
		gallery.InitGalleryRouter(group)
		order.InitOrderRouter(group)
		ecomm.InitECommerceRouter(group)
		product.InitProductRouter(group)
	}

	user.InitPermissionRouter(group)
	return Router
}
