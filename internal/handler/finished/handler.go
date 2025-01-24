package finished

import "github.com/gin-gonic/gin"

func InitFinishedAllRouter(router *gin.RouterGroup) {
	finishedRouter := router.Group("finished")

	InitFinishedRouter(finishedRouter)
	InitManageRouter(finishedRouter)
	InitFinishedStockRouter(finishedRouter)
}
