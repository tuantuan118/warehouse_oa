package service

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"strings"
	"time"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

func GetProductList(product *models.Product, pn, pSize int) (interface{}, error) {
	db := global.Db.Model(&models.Product{})

	if product.ID != 0 {
		db = db.Where("id = ?", product.ID)
	}
	if product.Name != "" {
		slice := strings.Split(product.Name, ";")
		db = db.Where("name in ?", slice)
	}

	db = db.Preload("FinishedManage")

	return Pagination(db, []models.Product{}, pn, pSize)
}

func GetProductById(id int) (*models.Product, error) {
	db := global.Db.Model(&models.Product{})

	data := &models.Product{}
	err := db.Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user does not exist")
	}

	return data, err
}

func SaveProduct(product *models.Product) (*models.Product, error) {
	err := IfProductByName(product.Name)
	if err != nil {
		return nil, err
	}

	finishedManage, err := GetFinishedManageById(product.FinishedManageId)
	if err != nil {
		return nil, err
	}

	today := time.Now().Format("20060102")
	total, err := getTodayProductCount()
	if err != nil {
		return nil, err
	}

	product.OrderNumber = fmt.Sprintf("QY%s%d", today, total+10001)
	product.FinishedManage = finishedManage
	err = global.Db.Model(&models.Product{}).Create(product).Error

	return product, err
}

func getTodayProductCount() (int64, error) {
	today := time.Now().Format("2006-01-02")

	var total int64
	err := global.Db.Model(&models.Product{}).Where(
		"DATE_FORMAT(add_time, '%Y-%m-%d') >= ?", today).Count(&total).Error

	return total, err
}

func UpdateProduct(product *models.Product) (*models.Product, error) {
	if product.ID == 0 {
		return nil, errors.New("id is 0")
	}
	_, err := GetProductById(product.ID)
	if err != nil {
		return nil, err
	}

	return product, global.Db.Updates(&product).Error
}

func DelProduct(id int, username string) error {
	if id == 0 {
		return errors.New("id is 0")
	}

	data, err := GetProductById(id)
	if err != nil {
		return err
	}
	if data == nil {
		return errors.New("user does not exist")
	}

	data.Operator = username
	data.IsDeleted = true
	err = global.Db.Updates(&data).Error
	if err != nil {
		return err
	}

	return global.Db.Delete(&data).Error
}

// GetProductFieldList 获取字段列表
func GetProductFieldList(field string) ([]string, error) {
	db := global.Db.Model(&models.Product{})
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

// IfProductByName 判断用户名是否已存在
func IfProductByName(name string) error {
	var count int64
	err := global.Db.Model(&models.Product{}).Where("name = ?",
		name).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("user name already exists")
	}

	return nil
}
