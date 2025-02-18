package service

import (
	"errors"
	"gorm.io/gorm"
	"strings"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

// GetIngredientsList 获取配料列表
func GetIngredientsList(ingredients *models.Ingredients, pn, pSize int) (interface{}, error) {
	db := global.Db.Model(&models.Ingredients{})

	if ingredients.Name != "" {
		slice := strings.Split(ingredients.Name, ";")
		db = db.Where("name in ?", slice)
	}

	return Pagination(db, []models.Ingredients{}, pn, pSize)
}

// GetIngredientsById 根据ID获取配料详情
func GetIngredientsById(id int) (*models.Ingredients, error) {
	db := global.Db.Model(&models.Ingredients{})

	data := &models.Ingredients{}
	err := db.Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("配料不存在")
	}

	return data, err
}

// GetIngredientsByName 根据名字获取配料详情
func GetIngredientsByName(name string) ([]int, error) {
	slice := strings.Split(name, ";")

	db := global.Db.Model(&models.Ingredients{})
	idList := make([]int, 0)
	err := db.Select("id").Where("name in ?", slice).Find(&idList).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("配料不存在")
	}

	return idList, err
}

// SaveIngredients 新增配料
func SaveIngredients(ingredients *models.Ingredients) (*models.Ingredients, error) {
	err := IfIngredientsByName(ingredients.Name)
	if err != nil {
		return nil, err
	}

	err = global.Db.Model(&models.Ingredients{}).Create(ingredients).Error

	return ingredients, err
}

// UpdateIngredients 修改配料
func UpdateIngredients(ingredients *models.Ingredients) (*models.Ingredients, error) {
	if ingredients.ID == 0 {
		return nil, errors.New("id is 0")
	}
	_, err := GetIngredientsById(ingredients.ID)
	if err != nil {
		return nil, err
	}

	return ingredients, global.Db.Select(
		"operator",
		"remark",
		"name",
		"is_calculate",
	).Updates(&ingredients).Error
}

// DelIngredients 删除配料
func DelIngredients(id int, username string) error {
	if id == 0 {
		return errors.New("id is 0")
	}

	data, err := GetIngredientsById(id)
	if err != nil {
		return err
	}
	if data == nil {
		return errors.New("配料不存在")
	}

	data.Operator = username
	err = global.Db.Updates(&data).Error
	if err != nil {
		return err
	}

	return global.Db.Delete(&data).Error
}

// GetIngredientsFieldList 获取字段列表
func GetIngredientsFieldList(field string) ([]string, error) {
	db := global.Db.Model(&models.Ingredients{})
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

// IfIngredientsByName 判断名称是否已存在
func IfIngredientsByName(name string) error {
	var count int64
	err := global.Db.Model(&models.Ingredients{}).Where("name = ?",
		name).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("配料名已存在")
	}

	return nil
}
