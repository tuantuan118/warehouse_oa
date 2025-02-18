package service

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"time"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

// GetProductionList 查询成品报工
func GetProductionList(production *models.FinishedProduction,
	begTime, endTime string, pn, pSize int) (interface{}, error) {

	db := global.Db.Model(&models.FinishedProduction{})
	db.Preload("Finished")

	if production.FinishedId > 0 {
		db = db.Where("finished_id = ?", production.FinishedId)
	}
	if production.Status > 0 {
		db = db.Where("status = ?", production.Status)
	}
	if begTime != "" && endTime != "" {
		db = db.Where("DATE_FORMAT(finish_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
	}

	return Pagination(db, []models.FinishedProduction{}, pn, pSize)
}

func GetFinishedConsumeList(production *models.FinishedProduction,
	begTime, endTime string,
	pn, pSize int) (interface{}, error) {

	db := global.Db.Model(&models.FinishedConsume{})

	if production.ID > 0 {
		db = db.Where("id = ?", production.ID)
	}
	if production.FinishedId > 0 {
		db = db.Where("finished_id = ?", production.FinishedId)
	}
	if production.Status > 0 {
		db = db.Where("status = ?", production.Status)
	}
	if begTime != "" && endTime != "" {
		db = db.Where("DATE_FORMAT(finish_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
	}

	return Pagination(db, []models.FinishedConsume{}, pn, pSize)
}

func GetProductionById(id int) (*models.FinishedProduction, error) {
	db := global.Db.Model(&models.FinishedProduction{})

	data := &models.FinishedProduction{}
	err := db.Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("成品报工不存在")
	}

	return data, err
}

func SaveProduction(production *models.FinishedProduction) (*models.Finished, error) {
	finished, err := GetFinishedById(production.FinishedId)
	if err != nil {
		return nil, err
	}
	production.Finished = finished
	production.Status = 1
	production.FinishTime = nil
	if production.FinishHour > 0 {
		et := time.Now().Add(time.Duration(production.FinishHour) * time.Hour)
		production.EstimatedTime = &et
	} else {
		// 4=已超时
		production.Status = 4
		et := time.Now()
		production.EstimatedTime = &et
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

	// 扣除配料库存
	for _, material := range finished.Material {
		err = DeductStock(tx, production,
			&models.IngredientStock{
				IngredientId: material.IngredientId,
				StockNum:     material.Quantity * float64(production.ActualAmount),
				StockUnit:    material.StockUnit,
			})
		if err != nil {
			return nil, err
		}
	}

	// 添加成品出入库信息

	err = tx.Model(&models.Finished{}).Create(&finished).Error
	if err != nil {
		logrus.Info(err)
		return nil, err
	}

	return finished, err
}

func UpdateProduction(finished *models.Finished) (*models.Finished, error) {
	if finished.ID == 0 {
		return nil, errors.New("id is 0")
	}
	oldData, err := GetFinishedById(finished.ID)
	if err != nil {
		return nil, err
	}

	if oldData.Status == 2 || oldData.Status == 3 {
		return nil, errors.New("finished has been finished, can not update")
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

	finished.ExpectAmount = 0
	finished.Name = ""
	finished.FinishedManage = nil

	err = tx.Updates(&finished).Error
	if err != nil {
		return nil, err
	}

	return finished, err
}

func VoidProduction(id int, username string) error {
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

	if data.Status == 2 || data.Status == 3 {
		return errors.New("finished has been finished, can not update")
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

	// 扣除配料库存
	finishedManage, err := GetFinishedManageById(data.FinishedManageId)
	if err != nil {
		return err
	}
	for _, material := range finishedManage.Material {
		err = FinishedSaveInBound(tx, &models.IngredientInBound{
			BaseModel: models.BaseModel{
				Operator: username,
			},
			IngredientId:     material.IngredientStock.IngredientId,
			StockNum:         material.Quantity * float64(data.ExpectAmount),
			StockUnit:        material.IngredientStock.StockUnit,
			StockUser:        username,
			StockTime:        time.Now(),
			OperationType:    "入库",
			Balance:          material.Quantity * float64(data.ExpectAmount),
			OperationDetails: fmt.Sprintf("报工生产【%s】作废重新入库", data.Name),
		})
		if err != nil {
			return err
		}
	}

	data.Operator = username
	data.Status = 3

	return tx.Updates(&data).Error
}

func FinishProduction(id, amount int, username string) error {
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

	if data.Status == 2 || data.Status == 3 {
		return errors.New("finished has been finished, can not update")
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

	data.Operator = username
	data.Status = 2
	data.ActualAmount = amount
	data.Ratio = (float64(data.ActualAmount) / float64(data.ExpectAmount)) * float64(100)
	ft := time.Now()
	data.FinishTime = &ft
	data.OperationDetails = fmt.Sprintf("生产完工")
	data.Balance = data.ActualAmount
	data.Price = data.Cost / float64(data.ActualAmount)

	err = SaveFinishedStockByInBound(tx, data)
	if err != nil {
		return err
	}

	return tx.Updates(&data).Error
}

func DelProduction(id int, username string) error {
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

	if data.Status == 2 || data.Status == 3 {
		return errors.New("finished has been finished, can not update")
	}

	data.Operator = username
	data.IsDeleted = true
	err = global.Db.Updates(&data).Error
	if err != nil {
		return err
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

	// 扣除配料库存
	finishedManage, err := GetFinishedManageById(data.FinishedManageId)
	if err != nil {
		return err
	}
	for _, material := range finishedManage.Material {
		err = FinishedSaveInBound(tx, &models.IngredientInBound{
			BaseModel: models.BaseModel{
				Operator: username,
			},
			IngredientId:     material.IngredientStock.IngredientId,
			StockNum:         material.Quantity * float64(data.ExpectAmount),
			StockUnit:        material.IngredientStock.StockUnit,
			StockUser:        username,
			StockTime:        time.Now(),
			OperationType:    "入库",
			Balance:          material.Quantity * float64(data.ExpectAmount),
			OperationDetails: fmt.Sprintf("报工生产【%s】删除重新入库", data.Name),
		})
		if err != nil {
			return err
		}
	}

	return tx.Delete(&data).Error
}

func ProductSaveProduction(tx *gorm.DB, finished *models.Finished) error {
	manage, err := GetFinishedManageById(finished.FinishedManageId)
	if err != nil {
		return err
	}

	finished.FinishedManage = manage

	err = SaveFinishedStockByInBound(tx, finished)
	if err != nil {
		return err
	}
	err = tx.Model(&models.Finished{}).Create(finished).Error

	return err
}

func UpdateProductionBalance(tx *gorm.DB, finishedId, pn int,
	amount int) (float64, error) {
	var cost float64
	data := make([]models.Finished, 0)
	db := global.Db.Model(&models.Finished{})
	db = db.Where("finished_manage_id = ?", finishedId)
	db = db.Where("balance > 0")
	db = db.Order("id asc").Limit(10).Offset((pn - 1) * 10)
	err := db.Find(&data).Error
	if err != nil {
		return 0, err
	}
	if len(data) == 0 {
		return 0, nil
	}

	for n, _ := range data {
		d := data[n]
		if amount >= d.Balance {
			cost = cost + d.Price*float64(d.Balance)
			amount = amount - d.Balance
			d.Balance = 0
		} else {
			cost = cost + d.Price*float64(amount)
			d.Balance = d.Balance - amount
			amount = 0
		}
		err = tx.Model(&models.Finished{}).
			Where("id = ?", d.ID).
			Update("balance", d.Balance).Error
		if err != nil {
			return 0, err
		}
		if amount == 0 {
			break
		}
	}
	if amount > 0 {
		c, err := UpdateFinishedBalance(tx, finishedId, pn+1, amount)
		if err != nil {
			return 0, err
		}
		cost = cost + c
	}

	return cost, nil
}

func GetProductionByFinishedId(finishedId int) ([]models.FinishedProduction, int64, error) {
	var err error
	var total int64
	var data []models.FinishedProduction

	db := global.Db.Model(&models.FinishedProduction{})
	db.Where("finished_id = ?", finishedId)
	if err = db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, errors.New("成品报工不存在")
	}

	err = db.Find(&data).Error

	return data, total, err
}
