package service

import (
	"errors"
	"gorm.io/gorm"
	"strings"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

// GetFinishedList 获取成品列表
func GetFinishedList(ids, name string, pn, pSize int) (interface{}, error) {
	db := global.Db.Model(&models.Finished{})

	if ids != "" {
		slice := strings.Split(ids, ";")
		db = db.Where("id in ?", slice)
	}
	if name != "" {
		slice := strings.Split(name, ";")
		db = db.Where("name in ?", slice)
	}

	return Pagination(db, []models.Finished{}, pn, pSize)
}

// GetFinishedIngredients 获取成品配料
func GetFinishedIngredients(id int) ([]map[string]interface{}, error) {
	if id == 0 {
		return nil, errors.New("id is 0")
	}
	db := global.Db
	productIngredient := make([]models.FinishedMaterial, 0)

	err := db.Where("finished_id = ?", id).Preload(
		"IngredientStock.Ingredient").Find(&productIngredient).Error
	if err != nil {
		return nil, err
	}

	requestData := make([]map[string]interface{}, 0)
	for _, v := range productIngredient {
		ingredient, err := GetIngredientsById(*v.IngredientStock.IngredientId)
		if err != nil {
			return nil, err
		}
		requestData = append(requestData, map[string]interface{}{
			"stock_id":  v.IngredientStock.ID,
			"name":      ingredient.Name,
			"quantity":  v.Quantity,
			"stockUnit": v.IngredientStock.StockUnit,
		})
	}

	return requestData, err
}

func GetFinishedById(id int) (*models.Finished, error) {
	db := global.Db.Model(&models.Finished{})

	data := &models.Finished{}
	err := db.Where("id = ?", id).Preload(
		"Material.IngredientStock").First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user does not exist")
	}

	return data, err
}

func SaveFinished(finished *models.Finished) (*models.Finished, error) {
	if finished.Material == nil || len(finished.Material) == 0 {
		return nil, errors.New("ingredients is empty")
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

	for _, material := range finished.Material {
		stock := new(models.IngredientStock)
		stock, err = GetStockById(material.IngredientId)
		if err != nil {
			return nil, err
		}
		material.IngredientStock = stock
	}

	err = global.Db.Model(&models.Finished{}).Create(&finished).Error
	if err != nil {
		return nil, err
	}

	return finished, err
}

func UpdateFinished(finished *models.Finished) (*models.Finished, error) {
	if finished.ID == 0 {
		return nil, errors.New("id is 0")
	}
	_, err := GetFinishedById(finished.ID)
	if err != nil {
		return nil, err
	}

	total, err := GetFinishedByStatus(finished.ID, 1)
	if err != nil {
		return nil, err
	}
	if total > 0 {
		return nil, errors.New("exist finished, can not update")
	}

	if finished.Material == nil || len(finished.Material) == 0 {
		return nil, errors.New("ingredients is empty")
	}

	err = RemoveIngredients(finished.ID)
	if err != nil {
		return nil, err
	}

	for _, material := range finished.Material {
		stock := new(models.IngredientStock)
		stock, err = GetStockById(material.IngredientId)
		if err != nil {
			return nil, err
		}
		material.IngredientStock = stock
	}

	return finished, global.Db.Updates(&finished).Error
}

func DelFinished(id int, username string) error {
	if id == 0 {
		return errors.New("id is 0")
	}

	data, err := GetFinishedById(id)
	if err != nil {
		return err
	}
	if data == nil {
		return errors.New("user does not exist")
	}

	total, err := GetFinishedByStatus(id, 1)
	if err != nil {
		return err
	}
	if total > 0 {
		return errors.New("exist finished, can not delete")
	}

	data.Operator = username
	data.IsDeleted = true
	err = global.Db.Updates(&data).Error
	if err != nil {
		return err
	}

	return global.Db.Delete(&data).Error
}

// GetFinishedFieldList 获取字段列表
func GetFinishedFieldList(field string) ([]string, error) {
	db := global.Db.Model(&models.Finished{})
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

func RemoveIngredients(manageId int) error {
	return global.Db.Model(&models.FinishedMaterial{}).Where(
		"finished_manage_id = ?", manageId).Delete(&models.FinishedMaterial{}).Error

}
