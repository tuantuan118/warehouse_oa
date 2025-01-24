package user

import (
	"github.com/gin-gonic/gin"
	"strconv"
	"warehouse_oa/internal/handler"
	"warehouse_oa/internal/models"
	"warehouse_oa/internal/service"
	"warehouse_oa/utils"
)

type User struct{}

var u User

func InitUserRouter(router *gin.RouterGroup) {
	userRouter := router.Group("user")

	userRouter.GET("list", u.list)
	userRouter.GET("fields", u.fields)
	userRouter.GET("getPermissions", u.getPermissions)
	userRouter.GET("getRoles", u.getRoles)
	userRouter.POST("update", u.update)
	userRouter.POST("delete", u.delete)
	userRouter.POST("changePassword", u.changePassword)
	userRouter.POST("setRoles", u.setRoles)
}

func (*User) list(c *gin.Context) {
	pn, pSize := utils.ParsePaginationParams(c)
	user := &models.User{
		Username: c.DefaultQuery("username", ""),
		Nickname: c.DefaultQuery("nickname", ""),
	}
	data, err := service.GetUserList(user, pn, pSize)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*User) update(c *gin.Context) {
	user := &models.User{}
	if err := c.ShouldBindJSON(user); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	user.Operator = c.GetString("userName")
	data, err := service.UpdateUser(user)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*User) delete(c *gin.Context) {
	user := &models.User{}
	if err := c.ShouldBindJSON(user); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	user.Operator = c.GetString("userName")
	err := service.DelUser(user.ID, user.Operator)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, nil)
}

func (*User) fields(c *gin.Context) {
	field := c.DefaultQuery("field", "")
	data, err := service.GetUserFieldList(field)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*User) changePassword(c *gin.Context) {
	var request struct {
		Id          int    `json:"id" binding:"required"`
		OldPassWord string `json:"oldPassWord" binding:"required"`
		NewPassWord string `json:"newPassWord" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	operator := c.GetString("userName")
	err := service.ChangePassword(request.Id, request.OldPassWord, request.NewPassWord, operator)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}
	handler.Success(c, nil)
}

func (*User) setRoles(c *gin.Context) {
	var request struct {
		Id  int   `json:"id" binding:"required"`
		Ids []int `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	operator := c.GetString("userName")
	err := service.SetRoles(request.Id, request.Ids, operator)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}
	handler.Success(c, nil)
}

func (*User) getPermissions(c *gin.Context) {
	idStr := c.GetString("userId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		handler.BadRequest(c, "id参数错误")
		return
	}
	data, err := service.GetRolePermissions(id)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*User) getRoles(c *gin.Context) {
	idStr := c.GetInt("userId")
	data, err := service.GetRoles(idStr)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}
