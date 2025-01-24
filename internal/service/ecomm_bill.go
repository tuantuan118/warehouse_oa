package service

import (
	"errors"
	"gorm.io/gorm"
	"mime/multipart"
	"strconv"
	"time"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
	"warehouse_oa/utils"
)

func GetECommBillList(eCommBill *models.ECommBill, pn, pSize int) (interface{}, error) {
	db := global.Db.Model(&models.ECommBill{})

	if eCommBill.Name != "" {
		db = db.Where("name = ?", eCommBill.Name)
	}

	return Pagination(db, []models.ECommBill{}, pn, pSize)
}

func GetECommBillById(id int) (*models.ECommBill, error) {
	db := global.Db.Model(&models.ECommBill{})

	data := &models.ECommBill{}
	err := db.Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user does not exist")
	}

	return data, err
}

func SaveECommBill(eCommBill *models.ECommBill) (*models.ECommBill, error) {
	err := IfECommBillByName(eCommBill.Name)
	if err != nil {
		return nil, err
	}

	err = global.Db.Model(&models.ECommBill{}).Create(eCommBill).Error

	return eCommBill, err
}

func UpdateECommBill(eCommBill *models.ECommBill) (*models.ECommBill, error) {
	if eCommBill.ID == 0 {
		return nil, errors.New("id is 0")
	}
	_, err := GetECommBillById(eCommBill.ID)
	if err != nil {
		return nil, err
	}

	return eCommBill, global.Db.Updates(&eCommBill).Error
}

func DelECommBill(id int, username string) error {
	if id == 0 {
		return errors.New("id is 0")
	}

	data, err := GetECommBillById(id)
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

// GetECommBillFieldList 获取字段列表
func GetECommBillFieldList(field string) ([]string, error) {
	db := global.Db.Model(&models.ECommBill{})
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

// IfECommBillByName 判断用户名是否已存在
func IfECommBillByName(name string) error {
	var count int64
	err := global.Db.Model(&models.ECommBill{}).Where("name = ?",
		name).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("user name already exists")
	}

	return nil
}

func UploadECommBill(file *multipart.FileHeader, username string) error {
	dataList, err := utils.UploadXlsx(file)
	if err != nil {
		return err
	}

	fastBillList := make([]models.ECommBill, 0)
	for _, data := range dataList {
		amount, err := strconv.Atoi(data["数量"])
		if err != nil {
			return err
		}
		deliveryTime, err := time.Parse("2006-01-02 15:04:05", data["发货时间"])
		if err != nil {
			return err
		}

		fastBillList = append(fastBillList, models.ECommBill{
			BaseModel: models.BaseModel{
				Operator: username,
				Remark:   data["备注"],
			},
			Name:           data["客户名称"],
			OrderNumber:    data["主订单编号"],
			TrackingNumber: data["快递单号"],
			Title:          data["商品标题"],
			Specification:  data["商品销售规格"],
			Amount:         amount,
			DeliveryTime:   deliveryTime,
		})
	}

	return global.Db.Updates(&fastBillList).Error
}
