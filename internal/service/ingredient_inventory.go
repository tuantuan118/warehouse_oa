package service

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

// GetStockList 获取库存列表
func GetStockList(name string, pn, pSize int) (interface{}, error) {
	db := global.Db.Model(&models.IngredientStock{})
	db = db.Select("ingredient_id, stock_unit, sum(stock_num) as stock_num, max(add_time) as add_time")
	db = db.Group("ingredient_id, stock_unit")

	if name != "" {
		idList, err := GetIngredientsByName(name)
		if err != nil {
			return nil, err
		}
		db = db.Where("ingredient_id in ?", idList)
	}
	db = db.Preload("Ingredient")

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	if pn != 0 && pSize != 0 {
		offset := (pn - 1) * pSize
		db = db.Order("add_time desc").Limit(pSize).Offset(offset)
	}

	var data []models.IngredientStock
	err := db.Find(&data).Error

	return map[string]interface{}{
		"data":       data,
		"pageNo":     pn,
		"pageSize":   pSize,
		"totalCount": total,
	}, err
}

// GetStockName 获取库存的名字和单位
func GetStockName() (interface{}, error) {
	var data []models.IngredientStock
	db := global.Db.Model(&models.IngredientStock{})
	db = db.Distinct("ingredient_id, stock_unit")
	db = db.Preload("Ingredient")
	err := db.Find(&data).Error

	return data, err
}

// IfStockByIdAndUnit 查询配料和单位存不存在 存在true 不存在 false
func IfStockByIdAndUnit(id int, unit int) (bool, error) {
	var total int64
	db := global.Db.Model(&models.IngredientStock{})
	err := db.Where("ingredient_id = ? and stock_unit = ?", id, unit).Count(&total).Error
	if err != nil {
		return true, err
	}
	if total == 0 {
		return false, nil
	} else {
		return true, nil
	}
}

func GetStockById(id int) (*models.IngredientStock, error) {
	db := global.Db.Model(&models.IngredientStock{})

	data := &models.IngredientStock{}
	err := db.Preload("Ingredient").Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user does not exist")
	}

	return data, err
}

func GetStockByIngredient(ingredientId, stockUnit int) (*models.IngredientStock, error) {
	db := global.Db.Model(&models.IngredientStock{})

	if ingredientId == 0 {
		return nil, errors.New("配料ID错误")
	}
	db = db.Where("ingredient_id = ?", ingredientId)
	db = db.Where("stock_unit = ?", stockUnit)
	db = db.Preload("Ingredient")

	data := &models.IngredientStock{}
	err := db.First(&data).Error

	return data, err
}

// SaveStockByInBound 通过入库保存库存
func SaveStockByInBound(db *gorm.DB, inBound *models.IngredientInBound) error {
	if inBound.IngredientId == nil && *inBound.IngredientId == 0 {
		return errors.New("配料ID错误")
	}
	if inBound.StockUnit == 0 {
		return errors.New("配料单位错误")
	}

	_, err := SaveStock(db, &models.IngredientStock{
		BaseModel: models.BaseModel{
			Operator: inBound.Operator,
		},
		IngredientId: inBound.IngredientId,
		InBoundId:    &inBound.ID,
		StockNum:     inBound.StockNum,
		StockUnit:    inBound.StockUnit,
	})

	return err
}

// SaveStock 保存库存
func SaveStock(db *gorm.DB, stock *models.IngredientStock) (*models.IngredientStock, error) {
	_, err := GetIngredientsById(*stock.IngredientId)
	if err != nil {
		return nil, err
	}

	err = db.Model(&models.IngredientStock{}).Create(&stock).Error

	return stock, err
}

func UpdateStockByInBound(db *gorm.DB, oldInBound *models.IngredientInBound) error {
	data := &models.IngredientStock{}
	var total int64

	db = db.Model(&models.IngredientStock{})
	db = db.Where("ingredient_id = ?", *oldInBound.IngredientId)
	if oldInBound.StockUnit == 0 {
		return errors.New("stock unit error")
	}
	db = db.Where("stock_unit = ?", oldInBound.StockUnit)

	var err error
	err = db.Count(&total).Error
	if err != nil {
		return err
	}
	if total == 0 {
		return errors.New("data does not exist")
	}
	err = db.First(&data).Error
	if err != nil {
		return err
	}

	data.StockNum += oldInBound.StockNum
	return global.Db.Select("stock_num").Updates(&data).Error
}

// DeductStock 扣除库存, 并且新增消耗表
func DeductStock(db *gorm.DB, production *models.FinishedProduction,
	ingredientStock *models.IngredientStock) error {

	var err error
	for {
		if ingredientStock.StockNum == 0 {
			break
		}

		stock := &models.IngredientStock{}
		err = db.Model(&models.IngredientStock{}).
			Where("ingredient_id = ?", *ingredientStock.IngredientId).
			Where("stock_unit = ?", ingredientStock.StockUnit).
			Where("stock_num > ?", 0).
			Order("add_time asc").First(&stock).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(fmt.Sprintf("id: %d 配料库存不足", *ingredientStock.IngredientId))
		}
		if err != nil {
			return err
		}

		logrus.Infoln(stock.StockNum)
		logrus.Infoln(ingredientStock.StockNum)

		if stock.StockNum > ingredientStock.StockNum {
			// 更新库存
			_, err = SaveConsume(db, &models.IngredientConsume{
				BaseModel: models.BaseModel{
					Operator: production.Operator,
				},
				FinishedId:       &production.FinishedId,
				IngredientId:     stock.IngredientId,
				ProductionId:     &production.ID,
				InBoundId:        stock.InBoundId,
				StockNum:         0 - ingredientStock.StockNum,
				StockUnit:        ingredientStock.StockUnit,
				OperationType:    false,
				OperationDetails: fmt.Sprintf("报工生产【%s】", production.Finished.Name),
			})

			stock.StockNum -= ingredientStock.StockNum
			err = db.Select("stock_num").Updates(&stock).Error
			if err != nil {
				return err
			}
			ingredientStock.StockNum = 0
		} else {
			_, err = SaveConsume(db, &models.IngredientConsume{
				BaseModel: models.BaseModel{
					Operator: production.Operator,
				},
				FinishedId:       &production.FinishedId,
				IngredientId:     stock.IngredientId,
				ProductionId:     &production.ID,
				InBoundId:        stock.InBoundId,
				StockNum:         0 - stock.StockNum,
				StockUnit:        ingredientStock.StockUnit,
				OperationType:    false,
				OperationDetails: fmt.Sprintf("报工生产【%s】", production.Finished.Name),
			})

			ingredientStock.StockNum -= stock.StockNum

			// 删除库存
			err = db.Delete(&stock).Error
			if err != nil {
				return err
			}
		}
	}
	return err
}

// DeductOrderAttach 扣除订单附加材料
func DeductOrderAttach(db *gorm.DB, order *models.Order,
	ingredientStock *models.IngredientStock) error {

	var err error
	for {
		if ingredientStock.StockNum == 0 {
			break
		}

		stock := &models.IngredientStock{}
		err = global.Db.Model(&models.IngredientStock{}).
			Where("ingredient_id = ?", *ingredientStock.IngredientId).
			Where("stock_unit = ?", ingredientStock.StockUnit).
			Where("stock_num > ?", 0).
			Order("add_time asc").First(&stock).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(fmt.Sprintf("id: %d 附加材料库存不足", *ingredientStock.IngredientId))
		}
		if err != nil {
			return err
		}

		logrus.Infoln(stock.StockNum)
		logrus.Infoln(ingredientStock.StockNum)

		if stock.StockNum > ingredientStock.StockNum {
			// 更新库存
			_, err = SaveConsume(db, &models.IngredientConsume{
				BaseModel: models.BaseModel{
					Operator: order.Operator,
				},
				IngredientId:     stock.IngredientId,
				InBoundId:        stock.InBoundId,
				OrderId:          &order.ID,
				StockNum:         0 - ingredientStock.StockNum,
				StockUnit:        ingredientStock.StockUnit,
				OperationType:    false,
				OperationDetails: fmt.Sprintf("订单【%s】附加材料", order.OrderNumber),
			})

			stock.StockNum -= ingredientStock.StockNum
			err = db.Select("stock_num").Updates(&stock).Error
			if err != nil {
				return err
			}
			ingredientStock.StockNum = 0
		} else {
			_, err = SaveConsume(db, &models.IngredientConsume{
				BaseModel: models.BaseModel{
					Operator: order.Operator,
				},
				IngredientId:     stock.IngredientId,
				InBoundId:        stock.InBoundId,
				OrderId:          &order.ID,
				StockNum:         0 - stock.StockNum,
				StockUnit:        ingredientStock.StockUnit,
				OperationType:    false,
				OperationDetails: fmt.Sprintf("订单【%s】附加材料", order.OrderNumber),
			})

			ingredientStock.StockNum -= stock.StockNum

			// 删除库存
			err = db.Delete(&stock).Error
			if err != nil {
				return err
			}
		}
	}
	return err
}

// DeductStockByInBound 根据inbound删除库存 (修改和删除配料入库时使用)
func DeductStockByInBound(db *gorm.DB, inBound *models.IngredientInBound) error {
	if inBound.IngredientId == nil && *inBound.IngredientId == 0 {
		return errors.New("配料ID错误")
	}
	if inBound.StockUnit == 0 {
		return errors.New("配料单位错误")
	}
	if inBound.UnitPrice == 0 {
		return errors.New("配料价格错误")
	}
	stock, err := GetStockByIngredient(*inBound.IngredientId, inBound.StockUnit)
	if err != nil {
		return err
	}
	if stock.StockNum-inBound.StockNum == 0 {
		// 删除 stock
		return db.Delete(&stock).Error
	} else if stock.StockNum-inBound.StockNum > 0 {
		// 扣除库存
		stock.StockNum -= inBound.StockNum
		return db.Updates(&stock).Error
	} else {
		// 报错
		return errors.New("库存不足")
	}
}
