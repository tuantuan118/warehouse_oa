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

// GetProductConsumeOutList 获取产品出入库列表
func GetProductConsumeOutList(ids, stockUnit, begTime, endTime string,
	inOrOut, pn, pSize int) (interface{}, error) {
	db := global.Db.Model(&models.ProductConsume{})
	totalDb := global.Db.Model(&models.ProductConsume{})
	db.Preload("Product")

	if ids != "" {
		idList := strings.Split(ids, ";")
		db = db.Where("product_id in ?", idList)
		totalDb = totalDb.Where("product_id in ?", idList)
	}
	if begTime != "" && endTime != "" {
		db = db.Where("DATE_FORMAT(add_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
		totalDb = totalDb.Where("DATE_FORMAT(add_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
	}
	if inOrOut == 1 {
		db = db.Where("stock_num > 0")
		totalDb = totalDb.Where("stock_num > 0")
	}
	if inOrOut == 2 {
		db = db.Where("stock_num < 0")
		totalDb = totalDb.Where("stock_num < 0")
	}
	var total int64
	if err := totalDb.Count(&total).Error; err != nil {
		return nil, err
	}

	if pn != 0 && pSize != 0 {
		offset := (pn - 1) * pSize
		db = db.Order("add_time desc").Limit(pSize).Offset(offset)
	}

	var data []models.ProductConsume
	err := db.Find(&data).Error

	return map[string]interface{}{
		"data":       data,
		"pageNo":     pn,
		"pageSize":   pSize,
		"totalCount": total,
	}, err
}

// GetInventorySum 获取产品出入库列表
func GetInventorySum(ids, stockUnit, begTime, endTime string,
	inOrOut int) (interface{}, error) {

	enterDb := global.Db.Model(&models.ProductConsume{})
	outDb := global.Db.Model(&models.ProductConsume{})

	if ids != "" {
		idList := strings.Split(ids, ";")
		enterDb = enterDb.Where("product_id in ?", idList)
		outDb = outDb.Where("product_id in ?", idList)
	}
	if begTime != "" && endTime != "" {
		enterDb = enterDb.Where("DATE_FORMAT(add_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
		outDb = outDb.Where("DATE_FORMAT(add_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
	}

	var enterNum, outNum float64
	err := enterDb.Where("stock_num >= 0").Select("IFNULL(SUM(stock_num), 0) AS stock_num").First(&enterNum).Error
	if err != nil {
		return nil, err
	}
	err = outDb.Where("stock_num <= 0").Select("IFNULL(SUM(stock_num), 0) AS stock_num").First(&outNum).Error
	if err != nil {
		return nil, err
	}

	if inOrOut == 1 {
		outNum = 0
	}
	if inOrOut == 2 {
		enterNum = 0
	}

	return map[string]float64{
		"enter": enterNum,
		"out":   outNum,
	}, err
}

func GetProductInventorAmount(id int) (interface{}, error) {
	db := global.Db.Model(&models.ProductInventory{})
	db = db.Select("sum(amount) as amount")
	db = db.Where("product_id =  ?", id)

	data := make(map[string]interface{})
	err := db.First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		data["id"] = id
		data["amount"] = 0
		return data, nil
	}
	if err != nil {
		return nil, err
	}
	data["id"] = id

	return data, err
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

	trueValue := true
	err = tx.Model(&models.ProductConsume{}).Create(&models.ProductConsume{
		BaseModel: models.BaseModel{
			Operator: data.Operator,
		},
		OrderId:          nil,
		ProductId:        data.ProductId,
		StockNum:         float64(data.Amount),
		OperationType:    &trueValue,
		OperationDetails: "添加产品",
	}).Error

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

	falseValue := false
	err = tx.Model(&models.ProductConsume{}).Create(&models.ProductConsume{
		BaseModel: models.BaseModel{
			Operator: inventory.Operator,
		},
		OrderId:          nil,
		ProductId:        inventory.ProductId,
		StockNum:         0 - float64(inventory.Amount),
		OperationType:    &falseValue,
		OperationDetails: "扣除产品",
	}).Error
	if err != nil {
		return err
	}

	amount := inventory.Amount
	for {
		if amount <= 0 {
			break
		}

		var data *models.ProductInventory
		err = tx.Model(&models.ProductInventory{}).
			Preload("Product").
			Preload("InventoryContent").
			Where("product_id = ? and amount > 0", inventory.ProductId).
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
			data.Amount = 0
			err = tx.Select("amount").Updates(&data).Error
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
func DeductProductStock(db *gorm.DB, productId, amount int) (int, error) {
	// 根据订单产品名查询产品
	product := &models.Product{}
	err := db.Model(&models.Product{}).Where(
		"id = ?", productId).First(&product).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return amount, nil
	}
	if err != nil {
		return 0, err
	}

	logrus.Infoln("11111111111111111111111111111111111")
	logrus.Infoln(product.ID)

	// 根据产品ID查询产品库存
	for amount > 0 {
		inventory := &models.ProductInventory{}
		err = db.Model(&models.ProductInventory{}).Where("product_id = ?", product.ID).First(&inventory).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return amount, nil
		}
		if err != nil {
			return 0, err
		}

		// 扣除库存

		if inventory.Amount-amount > 0 {
			inventory.Amount -= amount
			err = db.Select("amount").Updates(&inventory).Error
			if err != nil {
				return 0, err
			}
			amount = 0
		} else {
			amount -= inventory.Amount
			inventory.Amount = 0
			err = db.Select("amount").Updates(&inventory).Error
		}
	}
	return amount, nil
}
