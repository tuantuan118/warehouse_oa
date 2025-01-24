package service

import (
	"errors"
	"gorm.io/gorm"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

func GetCustomerList(customer *models.Customer, pn, pSize int) (interface{}, error) {
	db := global.Db.Model(&models.Customer{})

	if customer.Name != "" {
		db = db.Where("name = ?", customer.Name)
	}

	return Pagination(db, []models.Customer{}, pn, pSize)
}

func GetCustomerById(id int) (*models.Customer, error) {
	db := global.Db.Model(&models.Customer{})

	data := &models.Customer{}
	err := db.Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user does not exist")
	}

	return data, err
}

func SaveCustomer(customer *models.Customer) (*models.Customer, error) {
	err := IfCustomerByName(customer.Name)
	if err != nil {
		return nil, err
	}

	err = global.Db.Model(&models.Customer{}).Create(customer).Error

	return customer, err
}

func UpdateCustomer(customer *models.Customer) (*models.Customer, error) {
	if customer.ID == 0 {
		return nil, errors.New("id is 0")
	}
	_, err := GetCustomerById(customer.ID)
	if err != nil {
		return nil, err
	}

	return customer, global.Db.Updates(&customer).Error
}

func DelCustomer(id int, username string) error {
	if id == 0 {
		return errors.New("id is 0")
	}

	data, err := GetCustomerById(id)
	if err != nil {
		return err
	}
	if data == nil {
		return errors.New("user does not exist")
	}

	err = GetOrderByCustomer(id)
	if err != nil {
		return errors.New("existing orders")
	}

	data.Operator = username
	data.IsDeleted = true
	err = global.Db.Updates(&data).Error
	if err != nil {
		return err
	}

	return global.Db.Delete(&data).Error
}

// GetCustomerFieldList 获取字段列表
func GetCustomerFieldList(field string) ([]string, error) {
	db := global.Db.Model(&models.Customer{})
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

// IfCustomerByName 判断用户名是否已存在
func IfCustomerByName(name string) error {
	var count int64
	err := global.Db.Model(&models.Customer{}).Where("name = ?",
		name).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("user name already exists")
	}

	return nil
}
