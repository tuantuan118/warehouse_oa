package service

import (
	"errors"
	"github.com/go-sql-driver/mysql"
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
	total, err := IfProductByName(product.Name, product.Specification)
	if err != nil {
		return nil, err
	}
	if total > 0 {
		return nil, errors.New("产品名和规格已存在")
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

	err = global.Db.Model(&models.ProductInventory{}).Create(&models.ProductInventory{
		BaseModel: models.BaseModel{
			Operator: product.Operator,
		},
		ProductId:     product.ID,
		Amount:        0,
		ProductIdList: "",
	}).Error

	return product, err
}

// UpdateProduct 修改产品
func UpdateProduct(product *models.Product) (*models.Product, error) {
	var err error

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

	err = tx.Updates(&product).Error
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return nil, errors.New("产品名和规格已存在")
	}
	return product, err
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
func IfProductByName(name, specification string) (int64, error) {
	var total int64
	err := global.Db.Model(&models.Product{}).Where("name = ? and specification = ?",
		name, specification).Count(&total).Error
	if err != nil {
		return total, err
	}

	return total, nil
}

// GetProductByIndex 根据名字和规格查询产品
func GetProductByIndex(name, specification string) (*models.Product, error) {
	var product *models.Product
	err := global.Db.Model(&models.Product{}).Where("name = ? and specification = ?",
		name, specification).First(&product).Error

	return product, err
}

// RemoveFinished 删除关联的成品
func RemoveFinished(db *gorm.DB, productId int) error {
	return db.Model(&models.ProductContent{}).Where(
		"product_id = ?", productId).Delete(&models.ProductContent{}).Error

}
