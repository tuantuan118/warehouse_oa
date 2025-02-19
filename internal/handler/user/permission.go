package user

import (
	"github.com/gin-gonic/gin"
	"warehouse_oa/internal/handler"
	"warehouse_oa/internal/models"
	"warehouse_oa/internal/service"
	"warehouse_oa/utils"
)

type Permission struct{}

var p Permission

func InitPermissionRouter(router *gin.RouterGroup) {
	permissionRouter := router.Group("permission")

	permissionRouter.GET("list", p.list)
	permissionRouter.GET("fields", p.fields)
	permissionRouter.POST("add", p.add)
	permissionRouter.POST("update", p.update)
	permissionRouter.POST("delete", p.delete)
}

func (*Permission) list(c *gin.Context) {
	pn, pSize := utils.ParsePaginationParams(c)
	permission := &models.Permission{
		BaseModel: models.BaseModel{},
		Name:      c.DefaultQuery("name", ""),
		NameEn:    c.DefaultQuery("nameEn", ""),
		Coding:    c.DefaultQuery("coding", ""),
		Type:      utils.DefaultQueryInt(c, "type", 0),
		Enabled:   utils.DefaultQueryBool(c, "enabled", true),
	}
	data, err := service.GetPermissionList(permission, pn, pSize)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Permission) add(c *gin.Context) {
	permission := &models.Permission{}
	if err := c.ShouldBindJSON(permission); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	permission.Operator = c.GetString("userName")
	data, err := service.SavePermission(permission)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Permission) update(c *gin.Context) {
	permission := &models.Permission{}
	if err := c.ShouldBindJSON(permission); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	permission.Operator = c.GetString("userName")
	data, err := service.UpdatePermission(permission)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Permission) delete(c *gin.Context) {
	permission := &models.Permission{}
	if err := c.ShouldBindJSON(permission); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	err := service.DelPermission(permission.ID)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, nil)
}

func (*Permission) fields(c *gin.Context) {
	field := c.DefaultQuery("field", "")
	data, err := service.GetPermissionFieldList(field)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}
