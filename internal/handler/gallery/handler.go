package gallery

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"warehouse_oa/internal/handler"
	"warehouse_oa/internal/models"
	"warehouse_oa/internal/service"
	"warehouse_oa/utils"
)

type Gallery struct{}

var g Gallery

func InitGalleryRouter(router *gin.RouterGroup) {
	galleryRouter := router.Group("gallery")

	galleryRouter.GET("list", g.list)
	galleryRouter.GET("fields", g.fields)
	galleryRouter.POST("update", g.update)
	galleryRouter.POST("delete", g.delete)
	galleryRouter.POST("uploads", g.uploads)
}

func (*Gallery) list(c *gin.Context) {
	pn, pSize := utils.ParsePaginationParams(c)
	gallery := &models.Gallery{
		Name: c.DefaultQuery("name", ""),
	}
	data, err := service.GetGalleryList(gallery, pn, pSize)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Gallery) update(c *gin.Context) {
	gallery := &models.Gallery{}
	if err := c.ShouldBindJSON(gallery); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	gallery.Operator = c.GetString("userName")
	data, err := service.UpdateGallery(gallery)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Gallery) delete(c *gin.Context) {
	gallery := &models.Gallery{}
	if err := c.ShouldBindJSON(gallery); err != nil {
		// 如果解析失败，返回 400 错误和错误信息
		handler.BadRequest(c, err.Error())
		return
	}

	err := service.DelGallery(gallery.ID)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, nil)
}

func (*Gallery) fields(c *gin.Context) {
	field := c.DefaultQuery("field", "")
	data, err := service.GetGalleryFieldList(field)
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	handler.Success(c, data)
}

func (*Gallery) uploads(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 32<<20)
	
	form, err := c.MultipartForm()
	if err != nil {
		handler.InternalServerError(c, err)
		return
	}

	files := form.File["images"] // "images" 是表单的字段名
	if len(files) == 0 {
		handler.InternalServerError(c, errors.New("no files uploaded"))
		return
	}
	username := c.GetString("userName")

	m := map[string][]string{
		"success": {},
		"error":   {},
	}
	for _, file := range files {
		filename := file.Filename
		dst, name, err := service.SaveCosImages(file)
		if err != nil {
			m["error"] = append(m["error"], filename+err.Error())
			handler.InternalServerErrorData(c, err, m)
			return
		}

		if err = c.SaveUploadedFile(file, dst); err != nil {
			m["error"] = append(m["error"], filename+err.Error())
			handler.InternalServerErrorData(c, err, m)
			return
		}

		logrus.Infof("%s_%s_%s", dst, filename, username)
		err = service.SaveGallery(&models.Gallery{
			BaseModel: models.BaseModel{
				Operator: username,
			},
			Name: filename,
			Url:  name,
		})
		if err != nil {
			m["error"] = append(m["error"], filename+err.Error())
			handler.InternalServerErrorData(c, err, m)
			return
		}
		m["success"] = append(m["success"], filename)
	}

	handler.Success(c, m)
}
