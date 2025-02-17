package service

import (
	"errors"
	"gorm.io/gorm"
	"time"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

func GetStockList(name string, pn, pSize int) (interface{}, error) {
	db := global.Db.Model(&models.IngredientStock{})

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
	db = db.Where("unit_price = ?", unitPrice)
	db = db.Preload("Ingredient")

	data := &models.IngredientStock{}
	err := db.First(&data).Error

	return data, err
}

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
		UnitPrice:    inBound.UnitPrice,
		StockNum:     inBound.StockNum,
		StockUnit:    inBound.StockUnit,
	})

	return err
}

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

func DeductStock(db *gorm.DB, inBound *models.IngredientInBound) error {
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

// GetStockFieldList 获取字段列表
func GetStockFieldList(field string) (map[string]string, error) {
	db := global.Db.Model(&models.IngredientStock{})
	db = db.Select("id")
	switch field {
	case "name":
		db = db.Select("name")
	default:
		return nil, errors.New("field not exist")
	}
	fields := make(map[string]string)
	if err := db.Scan(&fields).Error; err != nil {
		return nil, err
	}

	return fields, nil
}
