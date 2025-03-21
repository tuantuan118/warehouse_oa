package service

import (
	"errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

// GetProductInventoryList 获取产品库存列表
func GetProductInventoryList(inventory *models.ProductInventory, pn, pSize int) (interface{}, error) {
	db := global.Db.Model(&models.ProductInventory{})

	if inventory.ID != 0 {
		db = db.Where("id = ?", inventory.ID)
	}
	if inventory.ProductId != 0 {
		db = db.Where("product_id = ?", inventory.ProductId)
	}

	db = db.Preload("Product")

	return Pagination(db, []models.ProductInventory{}, pn, pSize)
}

// GetProductInventoryById 根据ID获取产品库存
func GetProductInventoryById(id int) (*models.ProductInventory, error) {
	data := &models.ProductInventory{}
	err := global.Db.Model(&models.ProductInventory{}).
		Preload("Product").
		Where("id = ?", id).
		First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("产品库存不存在")
	}

	return data, err
}

// GetProductInventoryByProductId 根据产品ID获取库存
func GetProductInventoryByProductId(productId int) (*models.ProductInventory, error) {
	data := &models.ProductInventory{}
	err := global.Db.Model(&models.ProductInventory{}).
		Preload("Product").
		Where("product_id = ?", productId).
		First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("产品库存不存在")
	}

	return data, err
}

// SaveProductInventory 创建产品库存
func SaveProductInventory(data *models.ProductInventory) error {
	// 检查产品是否存在
	product, err := GetProductById(data.ProductId)
	if err != nil {
		return err
	}

	// 检查是否已存在该产品的库存记录
	exists, inventory, err := CheckProductInventoryExists(data.ProductId)
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

	if !exists {
		// 不存在
		err = db.Model(&models.ProductInventory{}).Create(data).Error
	} else {
		// 存在
		inventory.Amount += data.Amount
		err = db.Save(inventory).Error
	}

	// 消耗成品
	for _, u := range product.ProductContent {
		err = DeductFinishedStockByProduct(tx, product, &models.FinishedStock{
			FinishedId: u.FinishedId,
			Amount:     u.Quantity * float64(data.Amount),
		})
		if err != nil {
			return err
		}
	}
	return err
}

// UpdateProductInventory 更新产品库存
func UpdateProductInventory(inventory *models.ProductInventory) error {
	if inventory.Amount < 0 {
		return errors.New("参数错误")
	}

	// 获取现有记录
	existingInventory, err := GetProductInventoryById(inventory.ID)
	if err != nil {
		return err
	}

	// 更新记录
	existingInventory.Amount -= inventory.Amount
	existingInventory.Operator = inventory.Operator

	if existingInventory.Amount == 0 {
		err = global.Db.Delete(existingInventory).Error
	} else if existingInventory.Amount > 0 {
		err = global.Db.Select("amount", "operator").Updates(existingInventory).Error
	} else {
		return errors.New("库存数量不足扣除")
	}
	return err
}

// CheckProductInventoryExists 检查产品是否已有库存记录
func CheckProductInventoryExists(productId int) (bool, *models.ProductInventory, error) {
	var count int64
	data := &models.ProductInventory{}
	db := global.Db.Model(&models.ProductInventory{}).
		Where("product_id = ?", productId)

	err := db.Count(&count).Error
	if err != nil {
		return false, data, err
	}

	if count > 0 {
		err = db.First(&data).Error
		if err != nil {
			return false, data, err
		}
	}

	return count > 0, data, nil
}

// DeductProductStock 扣除产品库存
func DeductProductStock(db *gorm.DB, productName, specification string, amount int) (int, error) {
	// 根据订单产品名查询产品
	product := &models.Product{}
	err := db.Model(&models.Product{}).Where(
		"name = ? and specification = ?", productName, specification).First(&product).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return amount, nil
	}
	if err != nil {
		return 0, err
	}

	logrus.Infoln("11111111111111111111111111111111111")
	logrus.Infoln(product.ID)

	// 根据产品ID查询产品库存
	inventory := &models.ProductInventory{}
	err = db.Model(&models.ProductInventory{}).Where("product_id = ?", product.ID).First(&inventory).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return amount, nil
	}
	if err != nil {
		return 0, err
	}

	// 扣除库存
	inventory.Amount -= amount
	if inventory.Amount > 0 {
		err = db.Select("amount").Updates(&inventory).Error
		inventory.Amount = 0
	} else {
		err = db.Delete(&inventory).Error
	}

	return -inventory.Amount, err
}
