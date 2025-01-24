package service

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"strings"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

func GetInventoryList(name string, pn, pSize int) (interface{}, error) {
	db := global.Db.Model(&models.IngredientInventory{})

	if name != "" {
		idList, err := GetIngredientsByName(name)
		if err != nil {
			return nil, err
		}
		db = db.Where("ingredient_id in ?", idList)
	}
	db = db.Preload("Ingredient")

	return Pagination(db, []models.IngredientInventory{}, pn, pSize)
}

func GetInventoryById(id int) (*models.IngredientInventory, error) {
	db := global.Db.Model(&models.IngredientInventory{})

	data := &models.IngredientInventory{}
	err := db.Preload("Ingredient").Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user does not exist")
	}

	return data, err
}

// GetInventoryStockNumById 根据ID获取库存
//func GetInventoryStockNumById(ingredientId int) (*models.IngredientInventory, error) {
//	db := global.Db.Model(&models.IngredientInventory{})
//
//	data := &models.IngredientInventory{}
//	err := db.Where("ingredient_id = ?", ingredientId).First(&data).Error
//	if errors.Is(err, gorm.ErrRecordNotFound) {
//		return nil, errors.New("user does not exist")
//	}
//
//	return data, err
//}

func GetInventoryByIdList(ids string) ([]string, []string, error) {
	slice := strings.Split(ids, ";")

	db := global.Db.Model(&models.IngredientInventory{})
	data := make([]models.IngredientInventory, 0)
	err := db.Preload("Ingredient").Where("id in ?", slice).Find(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil, errors.New("user does not exist")
	}

	names := make([]string, 0)
	stockUnits := make([]string, 0)
	for _, v := range data {
		names = append(names, v.Ingredient.Name)
		stockUnits = append(stockUnits, fmt.Sprintf("%d", v.StockUnit))
	}

	return names, stockUnits, err
}

func SaveInventoryByInBound(db *gorm.DB, inBound *models.IngredientInBound) error {
	data := &models.IngredientInventory{}
	var total int64

	if inBound.StockUnit == 0 {
		return errors.New("stock unit error")
	}

	ingredientDb := global.Db.Model(&models.IngredientInventory{})
	ingredientDb = ingredientDb.Where("ingredient_id = ?", *inBound.IngredientID)

	ingredientDb = ingredientDb.Where("stock_unit = ?", inBound.StockUnit)
	var err error
	err = ingredientDb.Count(&total).Error
	if err != nil {
		return err
	}

	if total == 0 {
		_, err = SaveInventory(db, &models.IngredientInventory{
			BaseModel: models.BaseModel{
				Operator: inBound.Operator,
			},
			IngredientID:  inBound.IngredientID,
			Ingredient:    inBound.Ingredient,
			Specification: inBound.Specification,
			StockNum:      inBound.StockNum,
			StockUnit:     inBound.StockUnit,
		})
		return err
	}

	err = ingredientDb.First(&data).Error
	if err != nil {
		return err
	}

	data.StockNum += inBound.StockNum

	return db.Select("stock_num").Updates(&data).Error
}

func SaveInventory(db *gorm.DB, inventory *models.IngredientInventory) (*models.IngredientInventory, error) {
	ingredients, err := GetIngredientsById(*inventory.IngredientID)
	if err != nil {
		return nil, err
	}

	inventory.Ingredient = ingredients

	if inventory.StockNum < 0 {
		return nil, errors.New("insufficient inventory")
	}

	err = db.Model(&models.IngredientInventory{}).Create(&inventory).Error

	return inventory, err
}

func UpdateInventoryByInBound(db *gorm.DB, oldInBound *models.IngredientInBound) error {
	data := &models.IngredientInventory{}
	var total int64

	db = db.Model(&models.IngredientInventory{})
	db = db.Where("ingredient_id = ?", *oldInBound.IngredientID)
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

// GetInventoryFieldList 获取字段列表
func GetInventoryFieldList(field string) (map[string]string, error) {
	db := global.Db.Model(&models.IngredientInventory{})
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

//func UpdateIngredientStockNum(db *gorm.DB, id int, total int) error {
//	logrus.Infoln(total)
//	if id == 0 {
//		return errors.New("id is 0")
//	}
//
//	inventory, err := GetInventoryById(id)
//	if err != nil {
//		return err
//	}
//	if inventory.StockNum+total < 0 {
//		return errors.New("stock not enough")
//	}
//
//	inventory.StockNum += total
//
//	return db.Updates(&inventory).Error
//}
