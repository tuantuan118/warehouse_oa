package service

import (
	"errors"
	"gorm.io/gorm"
	"strings"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

func GetFinishedManageList(ids, name string, pn, pSize int) (interface{}, error) {
	db := global.Db.Model(&models.FinishedManage{})

	if ids != "" {
		slice := strings.Split(ids, ";")
		db = db.Where("id in ?", slice)
	}
	if name != "" {
		slice := strings.Split(name, ";")
		db = db.Where("name in ?", slice)
	}

	return Pagination(db, []models.FinishedManage{}, pn, pSize)
}

func GetFinishedManageIngredients(id int) ([]map[string]interface{}, error) {
	if id == 0 {
		return nil, errors.New("id is 0")
	}
	db := global.Db
	productIngredient := make([]models.FinishedMaterial, 0)

	err := db.Where("finished_manage_id = ?", id).Preload(
		"IngredientInventory.Ingredient").Find(&productIngredient).Error
	if err != nil {
		return nil, err
	}

	requestData := make([]map[string]interface{}, 0)
	for _, v := range productIngredient {
		ingredient, err := GetIngredientsById(*v.IngredientInventory.IngredientID)
		if err != nil {
			return nil, err
		}
		requestData = append(requestData, map[string]interface{}{
			"inventory_id": v.IngredientInventory.ID,
			"name":         ingredient.Name,
			"quantity":     v.Quantity,
			"stockUnit":    v.IngredientInventory.StockUnit,
		})
	}

	return requestData, err
}

func GetFinishedManageById(id int) (*models.FinishedManage, error) {
	db := global.Db.Model(&models.FinishedManage{})

	data := &models.FinishedManage{}
	err := db.Where("id = ?", id).Preload(
		"Material.IngredientInventory").First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user does not exist")
	}

	return data, err
}

func SaveFinishedManage(finishedManage *models.FinishedManage) (*models.FinishedManage, error) {
	if finishedManage.Material == nil || len(finishedManage.Material) == 0 {
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

	for _, material := range finishedManage.Material {
		inventory := new(models.IngredientInventory)
		inventory, err = GetInventoryById(material.IngredientID)
		if err != nil {
			return nil, err
		}
		material.IngredientInventory = inventory
	}

	err = global.Db.Model(&models.FinishedManage{}).Create(&finishedManage).Error
	if err != nil {
		return nil, err
	}

	return finishedManage, err
}

func UpdateFinishedManage(finishedManage *models.FinishedManage) (*models.FinishedManage, error) {
	if finishedManage.ID == 0 {
		return nil, errors.New("id is 0")
	}
	_, err := GetFinishedManageById(finishedManage.ID)
	if err != nil {
		return nil, err
	}

	total, err := GetFinishedByStatus(finishedManage.ID, 1)
	if err != nil {
		return nil, err
	}
	if total > 0 {
		return nil, errors.New("exist finished, can not update")
	}

	if finishedManage.Material == nil || len(finishedManage.Material) == 0 {
		return nil, errors.New("ingredients is empty")
	}

	err = RemoveIngredients(finishedManage.ID)
	if err != nil {
		return nil, err
	}

	for _, material := range finishedManage.Material {
		inventory := new(models.IngredientInventory)
		inventory, err = GetInventoryById(material.IngredientID)
		if err != nil {
			return nil, err
		}
		material.IngredientInventory = inventory
	}

	return finishedManage, global.Db.Updates(&finishedManage).Error
}

func DelFinishedManage(id int, username string) error {
	if id == 0 {
		return errors.New("id is 0")
	}

	data, err := GetFinishedManageById(id)
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

// GetFinishedManageFieldList 获取字段列表
func GetFinishedManageFieldList(field string) ([]string, error) {
	db := global.Db.Model(&models.FinishedManage{})
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
