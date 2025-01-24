package user

import (
	"github.com/gin-gonic/gin"
	"warehouse_oa/internal/handler"
	"warehouse_oa/internal/models"
	"warehouse_oa/internal/service"
)

type Login struct{}

var l Login

func InitLoginRouter(router *gin.RouterGroup) {
	userRouter := router.Group("user")

	userRouter.GET("ping", l.ping)
	userRouter.POST("login", l.login)
	userRouter.POST("register", l.register)
}

func (*Login) login(c *gin.Context) {
	user := &models.User{}
	if err := c.ShouldBindJSON(user); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	data, err := service.Login(user.Username, user.Password)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Login) register(c *gin.Context) {
	user := &models.User{}
	if err := c.ShouldBindJSON(user); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	data, err := service.Register(user)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Login) ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}
