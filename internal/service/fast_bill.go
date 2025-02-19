package service

import (
	"errors"
	"gorm.io/gorm"
	"mime/multipart"
	"strconv"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
	"warehouse_oa/utils"
)

func GetFastBillList(fastBill *models.FastBill, pn, pSize int) (interface{}, error) {
	db := global.Db.Model(&models.FastBill{})

	if fastBill.OrderNumber != "" {
		db = db.Where("order_number = ?", fastBill.OrderNumber)
	}

	return Pagination(db, []models.FastBill{}, pn, pSize)
}

func GetFastBillById(id int) (*models.FastBill, error) {
	db := global.Db.Model(&models.FastBill{})

	data := &models.FastBill{}
	err := db.Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user does not exist")
	}

	return data, err
}

func SaveFastBill(fastBill *models.FastBill) (*models.FastBill, error) {
	err := IfFastBillByName(fastBill.OrderNumber)
	if err != nil {
		return nil, err
	}

	err = global.Db.Model(&models.FastBill{}).Create(fastBill).Error

	return fastBill, err
}

func UpdateFastBill(fastBill *models.FastBill) (*models.FastBill, error) {
	if fastBill.ID == 0 {
		return nil, errors.New("id is 0")
	}
	_, err := GetFastBillById(fastBill.ID)
	if err != nil {
		return nil, err
	}

	return fastBill, global.Db.Updates(&fastBill).Error
}

func DelFastBill(id int) error {
	if id == 0 {
		return errors.New("id is 0")
	}

	data, err := GetFastBillById(id)
	if err != nil {
		return err
	}
	if data == nil {
		return errors.New("user does not exist")
	}

	return global.Db.Delete(&data).Error
}

// GetFastBillFieldList 获取字段列表
func GetFastBillFieldList(field string) ([]string, error) {
	db := global.Db.Model(&models.FastBill{})
	switch field {
	case "orderNumber":
		db.Select("order_number")
	default:
		return nil, errors.New("field not exist")
	}
	fields := make([]string, 0)
	if err := db.Scan(&fields).Error; err != nil {
		return nil, err
	}

	return fields, nil
}

// IfFastBillByName 判断订单编号是否已存在
func IfFastBillByName(name string) error {
	var count int64
	err := global.Db.Model(&models.FastBill{}).Where("order_number = ?",
		name).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("user name already exists")
	}

	return nil
}

func UploadFastBill(file *multipart.FileHeader, username string) error {
	dataList, err := utils.UploadXlsx(file)
	if err != nil {
		return err
	}

	fastBillList := make([]models.FastBill, 0)
	for _, data := range dataList {
		amount, err := strconv.Atoi(data["数量"])
		if err != nil {
			return err
		}
		status, err := strconv.Atoi(data["状态"])
		if err != nil {
			return err
		}
		payAmount, err := strconv.ParseFloat(data["赔付金额"], 64)
		if err != nil {
			return err
		}

		fastBillList = append(fastBillList, models.FastBill{
			BaseModel: models.BaseModel{
				Operator: username,
				Remark:   data["备注"],
			},
			OrderNumber:    data["订单编号"],
			TrackingNumber: data["快递单号"],
			Title:          data["商品标题"],
			Specification:  data["商品销售规格"],
			Amount:         amount,
			Status:         status,
			PayAmount:      payAmount,
		})
	}

	return global.Db.Updates(&fastBillList).Error
}
