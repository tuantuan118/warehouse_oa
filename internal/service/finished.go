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
func GetFinishedIngredients(id int) (interface{}, error) {
	if id == 0 {
		return nil, errors.New("id is 0")
	}
	db := global.Db
	productIngredient := make([]models.FinishedMaterial, 0)

	err := db.Where("finished_id = ?", id).Preload(
		"Ingredient").Find(&productIngredient).Error
	if err != nil {
		return nil, err
	}

	return productIngredient, err
}

// GetFinishedById ID查询成品
func GetFinishedById(id int) (*models.Finished, error) {
	db := global.Db.Model(&models.Finished{})

	data := &models.Finished{}
	err := db.Where("id = ?", id).Preload(
		"Material.Ingredient").First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("查找成品失败")
	}

	return data, err
}

// SaveFinished 新增成品
func SaveFinished(finished *models.Finished) (*models.Finished, error) {
	var err error

	for _, material := range finished.Material {
		_, err = GetIngredientsById(material.IngredientId)
		if err != nil {
			return nil, err
		}
	}

	err = global.Db.Model(&models.Finished{}).Create(&finished).Error
	if err != nil {
		return nil, err
	}

	return finished, err
}

// UpdateFinished 修改成品
func UpdateFinished(finished *models.Finished) (*models.Finished, error) {
	if finished.ID == 0 {
		return nil, errors.New("id is 0")
	}
	_, err := GetFinishedById(finished.ID)
	if err != nil {
		return nil, err
	}

	// 判断配料是否都存在
	for _, material := range finished.Material {
		material.Ingredient, err = GetIngredientsById(material.IngredientId)
		if err != nil {
			return nil, err
		}
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

	// 删除关联
	err = RemoveIngredients(tx, finished.ID)
	if err != nil {
		return nil, err
	}

	return finished, tx.Updates(&finished).Error
}

func DelFinished(id int) error {
	if id == 0 {
		return errors.New("id is 0")
	}

	data, err := GetFinishedById(id)
	if err != nil {
		return err
	}
	if data == nil {
		return errors.New("查询成品失败")
	}

	_, total, err := GetProductionByFinishedId(id)
	if err != nil {
		return err
	}
	if total > 0 {
		return errors.New("成品有报工记录，无法删除")
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

	// 删除关联
	err = RemoveIngredients(tx, data.ID)
	if err != nil {
		return err
	}

	return tx.Delete(&data).Error
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

// RemoveIngredients 删除关联的配料
func RemoveIngredients(db *gorm.DB, finishedId int) error {
	return db.Model(&models.FinishedMaterial{}).Where(
		"finished_id = ?", finishedId).Delete(&models.FinishedMaterial{}).Error
}
