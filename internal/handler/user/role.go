package user

import (
	"github.com/gin-gonic/gin"
	"warehouse_oa/internal/handler"
	"warehouse_oa/internal/models"
	"warehouse_oa/internal/service"
	"warehouse_oa/utils"
)

type Role struct{}

var r Role

func InitRoleRouter(router *gin.RouterGroup) {
	roleRouter := router.Group("role")

	roleRouter.GET("list", r.list)
	roleRouter.GET("fields", r.fields)
	roleRouter.POST("add", r.add)
	roleRouter.POST("update", r.update)
	roleRouter.POST("delete", r.delete)
	roleRouter.POST("setPermissions", r.setPermissions)
}

func (*Role) list(c *gin.Context) {
	pn, pSize := utils.ParsePaginationParams(c)
	role := &models.Role{
		BaseModel: models.BaseModel{},
		Name:      c.DefaultQuery("name", ""),
		Enabled:   utils.DefaultQueryBool(c, "enabled", true),
	}
	data, err := service.GetRoleList(role, pn, pSize)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Role) add(c *gin.Context) {
	role := &models.Role{}
	if err := c.ShouldBindJSON(role); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	role.Operator = c.GetString("userName")
	data, err := service.SaveRole(role)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Role) update(c *gin.Context) {
	role := &models.Role{}
	if err := c.ShouldBindJSON(role); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	role.Operator = c.GetString("userName")
	data, err := service.UpdateRole(role)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Role) delete(c *gin.Context) {
	role := &models.Role{}
	if err := c.ShouldBindJSON(role); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	err := service.DelRole(role.ID)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, nil)
}

func (*Role) fields(c *gin.Context) {
	field := c.DefaultQuery("field", "")
	data, err := service.GetRoleFieldList(field)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Role) setPermissions(c *gin.Context) {
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
	err := service.SetPermissions(request.Id, request.Ids, operator)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}
	handler.Success(c, nil)
}
