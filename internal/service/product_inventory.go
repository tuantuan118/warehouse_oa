package service

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"strings"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

// GetProductInventoryList 获取产品库存列表
func GetProductInventoryList(inventory *models.ProductInventory, pn, pSize int) (interface{}, error) {
	db := global.Db.Model(&models.ProductInventory{})

	db = db.Select("product_id, sum(amount) as amount, max(add_time) as add_time")
	db = db.Group("product_id")
	db.Preload("Product")

	if len(inventory.ProductIdList) != 0 {
		slice := strings.Split(inventory.ProductIdList, ";")
		db = db.Where("product_id in ?", slice)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	if pn != 0 && pSize != 0 {
		offset := (pn - 1) * pSize
		db = db.Order("add_time desc").Limit(pSize).Offset(offset)
	}

	var data []models.ProductInventory
	err := db.Find(&data).Error

	return map[string]interface{}{
		"data":       data,
		"pageNo":     pn,
		"pageSize":   pSize,
		"totalCount": total,
	}, err
}

// GetProductInventoryById 根据ID获取产品库存
func GetProductInventoryById(id int) (*models.ProductInventory, error) {
	data := &models.ProductInventory{}
	err := global.Db.Model(&models.ProductInventory{}).
		Preload("Product").
		Preload("InventoryContent").
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

	db := global.Db
	tx := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	for _, content := range product.ProductContent {
		data.InventoryContent = append(data.InventoryContent, models.InventoryContent{
			FinishedId: content.FinishedId,
			Quantity:   content.Quantity,
		})
	}

	fmt.Println(product.ProductContent)
	fmt.Println(data.InventoryContent)
	err = tx.Model(&models.ProductInventory{}).Create(data).Error
	if err != nil {
		return err
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

// UpdateProductInventory 扣除产品库存
func UpdateProductInventory(inventory *models.ProductInventory) error {
	if inventory.Amount < 0 {
		return errors.New("参数错误")
	}

	var err error
	db := global.Db
	tx := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	amount := inventory.Amount
	for {
		if amount <= 0 {
			break
		}

		var data *models.ProductInventory
		err = tx.Model(&models.ProductInventory{}).
			Preload("Product").
			Preload("InventoryContent").
			Where("product_id = ?", inventory.ProductId).
			Order("add_time asc").First(&data).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("库存不足")
		}
		if err != nil {
			return err
		}

		data.Operator = inventory.Operator
		err = ReturningInventory(tx, data, amount)
		if err != nil {
			return err
		}

		if data.Amount > amount {
			data.Amount -= amount
			amount = 0
			err = tx.Select("amount").Updates(&data).Error
			if err != nil {
				return err
			}
		} else {
			// 删除库存
			amount -= data.Amount
			err = tx.Model(&models.InventoryContent{}).
				Where("inventory_id = ?", data.ID).
				Delete(&models.InventoryContent{}).Error
			if err != nil {
				return err
			}

			err = tx.Delete(&data).Error
			if err != nil {
				return err
			}
		}
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
