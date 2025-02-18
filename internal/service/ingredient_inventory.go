package service

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"time"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

// GetStockList 获取库存列表
func GetStockList(name string, pn, pSize int) (interface{}, error) {
	db := global.Db.Model(&models.IngredientStock{})
	db = db.Select("ingredient_id, stock_unit, sum(stock_num) as stock_num")
	db = db.Group("ingredient_id, stock_unit")

	if name != "" {
		idList, err := GetIngredientsByName(name)
		if err != nil {
			return nil, err
		}
		db = db.Where("ingredient_id in ?", idList)
	}
	db = db.Preload("Ingredient")

	return Pagination(db, []models.IngredientStock{}, pn, pSize)
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

func GetStockById(id int) (*models.IngredientStock, error) {
	db := global.Db.Model(&models.IngredientStock{})

	data := &models.IngredientStock{}
	err := db.Preload("Ingredient").Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user does not exist")
	}

	return data, err
}

func GetStockByIngredient(ingredientId, stockUnit int, unitPrice float64) (*models.IngredientStock, error) {
	db := global.Db.Model(&models.IngredientStock{})

	if ingredientId == 0 {
		return nil, errors.New("配料ID错误")
	}
	db = db.Where("ingredient_id = ?", ingredientId)
	db = db.Where("stock_unit = ?", stockUnit)
	if unitPrice != 0 {
		db = db.Where("unit_price = ?", unitPrice)
	}
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
	if inBound.UnitPrice == 0 {
		return errors.New("配料价格错误")
	}
	if inBound.CreatedAt.IsZero() {
		inBound.CreatedAt = time.Now()
	}

	_, err := SaveStock(db, &models.IngredientStock{
		BaseModel: models.BaseModel{
			Operator:  inBound.Operator,
			CreatedAt: inBound.CreatedAt,
		},
		IngredientId: inBound.IngredientId,
		InBoundId:    &inBound.ID,
		InBound:      inBound,
		StockNum:     inBound.StockNum,
		StockUnit:    inBound.StockUnit,
	})

	return err
}

// SaveStock 保存库存
func SaveStock(db *gorm.DB, stock *models.IngredientStock) (*models.IngredientStock, error) {
	ingredients, err := GetIngredientsById(*stock.IngredientId)
	if err != nil {
		return nil, err
	}

	stock.Ingredient = ingredients

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
		stock := &models.IngredientStock{}
		err = global.Db.Model(&models.IngredientStock{}).
			Where("ingredient_id = ?", *ingredientStock.IngredientId).
			Where("stock_unit = ?", ingredientStock.StockUnit).
			Where("stock_num > ?", 0).
			Order("created_at asc").First(&stock).Error
		if err != nil {
			break
		}

		if stock.StockUnit > ingredientStock.StockUnit {
			// 更新库存
			_, err = SaveConsume(db, &models.IngredientConsume{
				FinishedId:   &production.FinishedId,
				Finish:       production.Finished,
				IngredientId: stock.IngredientId,
				//Ingredient:   stock.Ingredient,
				ProductionId: &production.ID,
				Production:   production,
				InBoundId:    stock.InBoundId,
				//InBound:          stock,
				StockNum:         ingredientStock.StockNum,
				StockUnit:        ingredientStock.StockUnit,
				OperationType:    false,
				OperationDetails: fmt.Sprintf("报工生产【%s】", production.Finished.Name),
			})

			stock.StockUnit -= ingredientStock.StockUnit
			err = db.Updates(&stock).Error
			if err != nil {
				return err
			}
		} else {
			_, err = SaveConsume(db, &models.IngredientConsume{
				FinishedId:   &production.FinishedId,
				Finish:       production.Finished,
				IngredientId: stock.IngredientId,
				//Ingredient:   stock.Ingredient,
				ProductionId: &production.ID,
				Production:   production,
				InBoundId:    stock.InBoundId,
				//InBound:          stock,
				StockNum:         ingredientStock.StockNum,
				StockUnit:        ingredientStock.StockUnit,
				OperationType:    false,
				OperationDetails: fmt.Sprintf("报工生产【%s】", production.Finished.Name),
			})

			// 删除库存
			err = db.Delete(&stock).Error
			if err != nil {
				return err
			}
			ingredientStock.StockUnit -= stock.StockUnit
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
	stock, err := GetStockByIngredient(*inBound.IngredientId, inBound.StockUnit, inBound.UnitPrice)
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
