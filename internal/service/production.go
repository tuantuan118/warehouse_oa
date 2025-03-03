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

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	if pn != 0 && pSize != 0 {
		offset := (pn - 1) * pSize
		db = db.Order("id desc").Limit(pSize).Offset(offset)
	}

	var data []models.FinishedProduction
	err := db.Find(&data).Error
	if err != nil {
		return nil, err
	}

	for _, item := range data {
		cost, err := GetCostByProduction(item.ID, 0)
		if err != nil {
			return nil, err
		}
		item.Cost = cost
	}

	return map[string]interface{}{
		"data":       data,
		"pageNo":     pn,
		"pageSize":   pSize,
		"totalCount": total,
	}, err
}

func GetFinishedConsumeList(production *models.FinishedProduction,
	begTime, endTime string,
	inOrOut, pn, pSize int) (interface{}, error) {

	db := global.Db.Model(&models.FinishedConsume{})
	db.Preload("Finished")

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
		db = db.Where("DATE_FORMAT(add_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
	}
	if inOrOut == 1 {
		db = db.Where("stock_num > 0")
	}
	if inOrOut == 2 {
		db = db.Where("stock_num < 0")
	}

	return Pagination(db, []models.FinishedConsume{}, pn, pSize)
}

func GetFinishedConsumeChart(production *models.FinishedProduction, begTime, endTime string) (interface{}, error) {
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
		db = db.Where("DATE_FORMAT(add_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
	}

	data := make([]map[string]interface{}, 0)
	err := db.Select("finished_id, stock_num").Find(&data).Error

	return data, err
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

func SaveProduction(production *models.FinishedProduction) (*models.FinishedProduction, error) {
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

	err = tx.Model(&models.FinishedProduction{}).Create(&production).Error
	if err != nil {
		logrus.Info(err)
		return nil, err
	}

	// 扣除配料库存

	for _, material := range finished.Material {
		logrus.Info("material.Quantity", material.Quantity)
		logrus.Info("production.ExpectAmount", production.ExpectAmount)

		err = DeductStock(tx, production,
			&models.IngredientStock{
				IngredientId: &material.IngredientId,
				StockUnit:    material.StockUnit,
				StockNum:     material.Quantity * float64(production.ExpectAmount),
			})
		if err != nil {
			return nil, err
		}
	}

	return production, err
}

// VoidProduction 作废报工
func VoidProduction(id int, username string) error {
	if id == 0 {
		return errors.New("id is 0")
	}

	production, err := GetProductionById(id)
	if err != nil {
		return err
	}
	if production == nil {
		return errors.New("报工单不存在")
	}
	production.Finished, err = GetFinishedById(production.FinishedId)
	if err != nil {
		return err
	}

	if production.Status == 2 || production.Status == 3 {
		return errors.New("已完工或以作废，无法修改")
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

	// 返还配料库存
	consumeList, err := GetConsumeByProduction(production.ID)
	if err != nil {
		return err
	}
	for _, consume := range consumeList {
		// 添加配料库存
		inBound := &models.IngredientInBound{
			BaseModel: models.BaseModel{
				ID:       *consume.InBoundId,
				Operator: username,
			},
			IngredientId: consume.IngredientId,
			StockUnit:    consume.StockUnit,
			StockNum:     -consume.StockNum,
		}
		err = SaveStockByInBound(tx, inBound)
		if err != nil {
			return err
		}

		// 添加配料消耗表
		err = SaveConsumeByInBound(tx, inBound,
			fmt.Sprintf("报工生产【%s】作废重新入库", production.Finished.Name))
	}

	production.Operator = username
	production.Status = 3

	return tx.Updates(&production).Error
}

// FinishProduction 完成报工
func FinishProduction(id, amount int, username string) error {
	if id == 0 {
		return errors.New("id is 0")
	}

	production, err := GetProductionById(id)
	if err != nil {
		return err
	}
	if production == nil {
		return errors.New("报工单不存在")
	}
	if production.Status == 2 || production.Status == 3 {
		return errors.New("已完工或以作废，无法修改")
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

	production.Operator = username
	production.Status = 2
	production.ActualAmount = amount
	production.Ratio = (float64(production.ActualAmount) / float64(production.ExpectAmount)) * float64(100)
	ft := time.Now()
	production.FinishTime = &ft

	// 添加成品库存
	err = SaveStockByProduction(tx, production)
	if err != nil {
		return err
	}

	// 添加成品出入库信息
	err = SaveConsumeByProduction(tx, production)
	if err != nil {
		return err
	}

	return tx.Updates(&production).Error
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
		return nil, 0, nil
	}

	err = db.Find(&data).Error

	return data, total, err
}
