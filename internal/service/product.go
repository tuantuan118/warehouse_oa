package service

import (
	"errors"
	"gorm.io/gorm"
	"strings"
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

	db = db.Preload("ProductContent.Finished")

	return Pagination(db, []models.Product{}, pn, pSize)
}

func GetProductById(id int) (*models.Product, error) {
	db := global.Db.Model(&models.Product{})
	db = db.Preload("ProductContent")

	data := &models.Product{}
	err := db.Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("产品不存在")
	}

	return data, err
}

func GetProductByName(name string) (*models.Product, error) {
	db := global.Db.Model(&models.Product{})
	db = db.Preload("ProductContent")

	data := &models.Product{}
	err := db.Where("name = ?", name).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("产品不存在")
	}

	return data, err
}

// SaveProduct 创建产品
func SaveProduct(product *models.Product) (*models.Product, error) {
	err := IfProductByName(product.Name)
	if err != nil {
		return nil, err
	}

	if product.ProductContent == nil || len(product.ProductContent) == 0 {
		return nil, errors.New("产品列表不能为空")
	}

	// 判断配料是否都存在
	for _, c := range product.ProductContent {
		c.Finished, err = GetFinishedById(c.FinishedId)
		if err != nil {
			return nil, err
		}
	}

	err = global.Db.Model(&models.Product{}).Create(product).Error

	return product, err
}

// UpdateProduct 修改产品
func UpdateProduct(product *models.Product) (*models.Product, error) {
	if product.ID == 0 {
		return nil, errors.New("id is 0")
	}
	_, err := GetProductById(product.ID)
	if err != nil {
		return nil, err
	}

	if product.ProductContent == nil || len(product.ProductContent) == 0 {
		return nil, errors.New("产品列表不能为空")
	}

	// 判断成品是否都存在
	for _, c := range product.ProductContent {
		c.Finished, err = GetFinishedById(c.FinishedId)
		if err != nil {
			return nil, err
		}
	}

	db := global.Db
	tx := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// 删除关联
	err = RemoveFinished(tx, product.ID)
	if err != nil {
		return nil, err
	}

	return product, tx.Updates(&product).Error
}

func DelProduct(id int) error {
	if id == 0 {
		return errors.New("id is 0")
	}

	data, err := GetProductById(id)
	if err != nil {
		return err
	}

	db := global.Db
	tx := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// 删除关联
	err = RemoveFinished(tx, id)
	if err != nil {
		return err
	}

	return tx.Delete(&data).Error
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
		return errors.New("产品名已存在")
	}

	return nil
}

// RemoveFinished 删除关联的成品
func RemoveFinished(db *gorm.DB, productId int) error {
	return db.Model(&models.ProductContent{}).Where(
		"product_id = ?", productId).Delete(&models.ProductContent{}).Error

}
