package service

import (
	"errors"
	"gorm.io/gorm"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

func GetECommCustomersList(eCommECommCustomers *models.ECommCustomers, pn, pSize int) (interface{}, error) {
	db := global.Db.Model(&models.ECommCustomers{})

	if eCommECommCustomers.Name != "" {
		db = db.Where("name = ?", eCommECommCustomers.Name)
	}

	return Pagination(db, []models.ECommCustomers{}, pn, pSize)
}

func GetECommCustomersById(id int) (*models.ECommCustomers, error) {
	db := global.Db.Model(&models.ECommCustomers{})

	data := &models.ECommCustomers{}
	err := db.Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user does not exist")
	}

	return data, err
}

func SaveECommCustomers(eCommECommCustomers *models.ECommCustomers) (*models.ECommCustomers, error) {
	err := IfECommCustomersByName(eCommECommCustomers.Name)
	if err != nil {
		return nil, err
	}

	err = global.Db.Model(&models.ECommCustomers{}).Create(eCommECommCustomers).Error

	return eCommECommCustomers, err
}

func UpdateECommCustomers(eCommECommCustomers *models.ECommCustomers) (*models.ECommCustomers, error) {
	if eCommECommCustomers.ID == 0 {
		return nil, errors.New("id is 0")
	}
	_, err := GetECommCustomersById(eCommECommCustomers.ID)
	if err != nil {
		return nil, err
	}

	return eCommECommCustomers, global.Db.Updates(&eCommECommCustomers).Error
}

func DelECommCustomers(id int, username string) error {
	if id == 0 {
		return errors.New("id is 0")
	}

	data, err := GetECommCustomersById(id)
	if err != nil {
		return err
	}
	if data == nil {
		return errors.New("user does not exist")
	}

	data.Operator = username
	data.IsDeleted = true
	err = global.Db.Updates(&data).Error
	if err != nil {
		return err
	}

	return global.Db.Delete(&data).Error
}

// GetECommCustomersFieldList 获取字段列表
func GetECommCustomersFieldList(field string) ([]string, error) {
	db := global.Db.Model(&models.ECommCustomers{})
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

// IfECommCustomersByName 判断用户名是否已存在
func IfECommCustomersByName(name string) error {
	var count int64
	err := global.Db.Model(&models.ECommCustomers{}).Where("name = ?",
		name).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("user name already exists")
	}

	return nil
}
