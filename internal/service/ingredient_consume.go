package service

import (
	"errors"
	"gorm.io/gorm"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

// GetConsumeList 返回出入库列表查询数据
func GetConsumeList(name, stockUnit, begTime, endTime string,
	pn, pSize int) (interface{}, error) {

	db := global.Db.Model(&models.IngredientConsume{})
	totalDb := global.Db.Model(&models.IngredientConsume{})

	if name != "" {
		idList, err := GetIngredientsByName(name)
		if err != nil {
			return nil, err
		}
		db = db.Where("ingredient_id in ?", idList)
		totalDb = totalDb.Where("ingredient_id in ?", idList)
	}
	if stockUnit != "" {
		db = db.Where("stock_unit = ?", stockUnit)
		totalDb = totalDb.Where("stock_unit = ?", stockUnit)
	}
	if begTime != "" && endTime != "" {
		db = db.Where("DATE_FORMAT(stock_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
		totalDb = totalDb.Where("DATE_FORMAT(stock_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
	}

	var cost float64
	if err := totalDb.Select("SUM(cost)").Scan(&cost).Error; err != nil {
		return nil, err
	}

	db = db.Preload("Ingredient")

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	if pn != 0 && pSize != 0 {
		offset := (pn - 1) * pSize
		db = db.Order("id desc").Limit(pSize).Offset(offset)
	}

	data := make([]models.IngredientConsume, 0)
	err := db.Find(&data).Error
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"data":       data,
		"pageNo":     pn,
		"pageSize":   pSize,
		"totalCount": total,
		"cost":       cost,
	}, err

}

// SaveConsumeByInBound 通过入库表来保存消耗表
func SaveConsumeByInBound(db *gorm.DB, inBound *models.IngredientInBound) error {
	if inBound.IngredientId == nil && *inBound.IngredientId == 0 {
		return errors.New("配料ID错误")
	}
	if inBound.ID == 0 {
		return errors.New("配料入库ID错误")
	}
	if inBound.StockUnit == 0 {
		return errors.New("配料单位错误")
	}
	if inBound.UnitPrice == 0 {
		return errors.New("配料价格错误")
	}

	_, err := SaveConsume(db, &models.IngredientConsume{
		BaseModel: models.BaseModel{
			Operator: inBound.Operator,
		},
		IngredientId:     inBound.IngredientId,
		InBoundId:        &inBound.ID,
		StockNum:         inBound.StockNum,
		StockUnit:        inBound.StockUnit,
		OperationType:    true,
		OperationDetails: "配料入库",
	})

	return err
}

// SaveConsume 保存消耗表
func SaveConsume(db *gorm.DB, consume *models.IngredientConsume) (*models.IngredientConsume, error) {
	ingredients, err := GetIngredientsById(*consume.IngredientId)
	if err != nil {
		return nil, err
	}

	inBound, err := GetInBoundById(*consume.InBoundId)
	if err != nil {
		return nil, err
	}

	consume.Ingredient = ingredients
	consume.InBound = inBound

	err = db.Model(&models.IngredientConsume{}).Create(&consume).Error

	return consume, err
}

// DelConsumeByInBound 通过入库表来删除消耗表
func DelConsumeByInBound(db *gorm.DB, id int) error {
	var total int64
	if err := global.Db.Model(&models.IngredientConsume{}).
		Where("in_bound_id = ? and operation_type = ?", id, false).Count(&total).Error; err != nil {
		return errors.New("配料已使用，无法删除")
	}

	err := db.Where("in_bound_id = ? and operation_type = ?", id, true).
		Delete(&models.IngredientConsume{}).Error

	return err
}

// UpdateInBoundBalance 更改余量接口（改到消耗表）
func UpdateInBoundBalance(tx *gorm.DB, stock *models.IngredientStock, pn int,
	amount float64) (float64, error) {

	//var cost float64
	//data := make([]models.IngredientInBound, 0)
	//db := global.Db.Model(&models.IngredientInBound{})
	//db = db.Where("ingredient_id = ?", stock.IngredientId)
	//db = db.Where("stock_unit = ?", stock.StockUnit)
	//db = db.Where("balance > 0")
	//db = db.Order("id asc").Limit(10).Offset((pn - 1) * 10)
	//err := db.Find(&data).Error
	//if err != nil {
	//	return 0, err
	//}
	//if len(data) == 0 {
	//	return 0, nil
	//}
	//
	//for n, _ := range data {
	//	d := &data[n]
	//
	//	if amount >= d.Balance {
	//		cost = cost + d.Price*d.Balance
	//		amount = amount - d.Balance
	//		d.Balance = 0
	//	} else {
	//		cost = cost + d.Price*amount
	//		d.Balance = d.Balance - amount
	//		amount = 0
	//	}
	//
	//	err = tx.Model(&models.IngredientInBound{}).
	//		Where("id = ?", d.ID).
	//		Update("balance", d.Balance).Error
	//	if err != nil {
	//		return 0, err
	//	}
	//	if amount == 0 {
	//		continue
	//	}
	//}
	//if amount > 0 {
	//	c, err := UpdateInBoundBalance(tx, stock, pn+1, amount)
	//	if err != nil {
	//		return 0, err
	//	}
	//	cost = cost + c
	//}
	//
	//return cost, nil
}

// FinishedSaveInBound 成品调用接口 （应该改到消耗表）
func FinishedSaveInBound(tx *gorm.DB, inBound *models.IngredientInBound) error {
	//ingredients, err := GetIngredientsById(*inBound.IngredientId)
	//if err != nil {
	//	return err
	//}
	//
	//inBound.Ingredient = ingredients
	//
	//err = SaveStockByInBound(tx, inBound)
	//if err != nil {
	//	return err
	//}
	//err = tx.Model(&models.IngredientInBound{}).Create(inBound).Error
	//
	//return nil
}
