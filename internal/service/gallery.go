package service

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"mime/multipart"
	"os"
	"path/filepath"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

func GetGalleryList(gallery *models.Gallery, pn, pSize int) (interface{}, error) {
	db := global.Db.Model(&models.Gallery{})

	if gallery.Name != "" {
		db = db.Where("name LIKE ?", "%"+gallery.Name+"%")
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	if pn != 0 && pSize != 0 {
		offset := (pn - 1) * pSize
		db = db.Order("id desc").Limit(pSize).Offset(offset)
	}

	data := make([]models.Gallery, 0)
	err := db.Find(&data).Error
	if err != nil {
		return nil, err
	}

	var imageUrls []map[string]interface{}
	for _, d := range data {
		imageUrls = append(imageUrls, map[string]interface{}{
			"id":   d.ID,
			"name": d.Name,
			"urls": fmt.Sprintf("http://8.138.155.131:8080/images/%s", d.Url),
		})
	}

	return imageUrls, err
}

func GetGalleryById(id int) (*models.Gallery, error) {
	db := global.Db.Model(&models.Gallery{})

	data := &models.Gallery{}
	err := db.Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user does not exist")
	}

	return data, err
}

func SaveGallery(gallery *models.Gallery) error {
	err := IfGalleryByName(gallery.Name)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = os.Remove(gallery.Url)
		}
	}()

	err = global.Db.Model(&models.Gallery{}).Create(gallery).Error

	return err
}

func UpdateGallery(gallery *models.Gallery) (*models.Gallery, error) {
	if gallery.ID == 0 {
		return nil, errors.New("id is 0")
	}
	_, err := GetGalleryById(gallery.ID)
	if err != nil {
		return nil, err
	}

	gallery.Url = ""

	return gallery, global.Db.Updates(&gallery).Error
}

func DelGallery(id int) error {
	if id == 0 {
		return errors.New("id is 0")
	}

	data, err := GetGalleryById(id)
	if err != nil {
		return err
	}
	if data == nil {
		return errors.New("user does not exist")
	}

	saveDir := "./cos/images"
	join := filepath.Join(saveDir, data.Url)
	err = os.Remove(join)
	if err != nil {
		return err
	}

	return global.Db.Delete(&data).Error
}

// GetGalleryFieldList 获取字段列表
func GetGalleryFieldList(field string) ([]string, error) {
	db := global.Db.Model(&models.Gallery{})
	switch field {
	case "name":
		db.Select("name")
	default:
		return nil, errors.New("field not exist")
	}
	fields := make([]string, 0)
	if err := db.Scan(&fields).Error; err != nil {
		return nil, err
	}

	return fields, nil
}

// IfGalleryByName 判断用户名是否已存在
func IfGalleryByName(name string) error {
	var count int64
	err := global.Db.Model(&models.Gallery{}).Where("name = ?",
		name).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("user name already exists")
	}

	return nil
}

func SaveCosImages(f *multipart.FileHeader) (string, string, error) {

	// 创建保存路径
	saveDir := "./cos/images"
	if _, err := os.Stat(saveDir); os.IsNotExist(err) {
		err = os.MkdirAll(saveDir, os.ModePerm)
		if err != nil {
			return "", "", err
		}
	}

	// 生成唯一文件名
	filename := fmt.Sprintf("%s%s", uuid.New().String(), filepath.Base(f.Filename))
	join := filepath.Join(saveDir, filename)

	return join, filename, nil
}
