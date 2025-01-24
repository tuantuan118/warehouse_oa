package ecomm

import "github.com/gin-gonic/gin"

func InitECommerceRouter(router *gin.RouterGroup) {
	eCommerceRouter := router.Group("e_commerce")

	InitECommBillRouter(eCommerceRouter)
	InitECommCustomersRouter(eCommerceRouter)
	InitFastBillRouter(eCommerceRouter)
}
