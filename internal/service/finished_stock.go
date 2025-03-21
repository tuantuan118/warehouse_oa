package service

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"strings"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

// GetFinishedStockList 查询库存列表接口
func GetFinishedStockList(ids string, begReportingTime, endReportingTime string,
	pn, pSize int) (interface{}, error) {

	db := global.Db.Model(&models.FinishedStock{})
	db = db.Select("finished_id, sum(amount) as amount, max(add_time) as add_time")
	db = db.Group("finished_id")
	db.Preload("Finished")

	if ids != "" {
		idList := strings.Split(ids, ";")
		db = db.Where("finished_id in ?", idList)
	}
	if begReportingTime != "" && endReportingTime != "" {
		db = db.Where("add_time BETWEEN ? AND ?", begReportingTime, endReportingTime)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	if pn != 0 && pSize != 0 {
		offset := (pn - 1) * pSize
		db = db.Order("add_time desc").Limit(pSize).Offset(offset)
	}

	var data []models.FinishedStock
	err := db.Find(&data).Error

	return map[string]interface{}{
		"data":       data,
		"pageNo":     pn,
		"pageSize":   pSize,
		"totalCount": total,
	}, err
}

// GetFinishedStockById 通过ID获取库存
func GetFinishedStockById(id int) (*models.FinishedStock, error) {
	db := global.Db.Model(&models.FinishedStock{})

	data := &models.FinishedStock{}
	err := db.Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user does not exist")
	}

	return data, err
}

// SaveStockByProduction 通过报工保存库存
func SaveStockByProduction(db *gorm.DB, production *models.FinishedProduction) error {
	if production.FinishedId == 0 {
		return errors.New("成品ID错误")
	}

	_, err := SaveFinishedStock(db, &models.FinishedStock{
		BaseModel: models.BaseModel{
			Operator: production.Operator,
		},
		FinishedId:   production.FinishedId,
		ProductionId: production.ID,
		Amount:       float64(production.ActualAmount),
	})

	return err
}

// SaveFinishedStock 保存成品库存
func SaveFinishedStock(db *gorm.DB, finished *models.FinishedStock) (*models.FinishedStock, error) {
	if finished.Amount < 0 {
		return nil, errors.New("成品数量错误")
	}

	err := db.Model(&models.FinishedStock{}).Create(&finished).Error

	return finished, err
}

// DeductFinishedStock 订单扣除库存, 并且新增消耗表
func DeductFinishedStock(db *gorm.DB, order *models.Order,
	finishedStock *models.FinishedStock) error {

	var err error
	for {
		if finishedStock.Amount == 0 {
			break
		}

		stock := &models.FinishedStock{}
		err = db.Model(&models.FinishedStock{}).
			Where("finished_id = ?", finishedStock.FinishedId).
			Where("amount > ?", 0).
			Order("add_time asc").First(&stock).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(fmt.Sprintf("id: %d 成品库存不足", finishedStock.FinishedId))
		}
		if err != nil {
			return err
		}

		if stock.Amount > finishedStock.Amount {
			// 更新库存
			_, err = SaveFinishedConsume(db, &models.FinishedConsume{
				BaseModel: models.BaseModel{
					Operator: order.Operator,
				},
				OrderId:          &order.ID,
				FinishedId:       stock.FinishedId,
				ProductionId:     stock.ProductionId,
				StockNum:         0 - finishedStock.Amount,
				OperationType:    false,
				OperationDetails: fmt.Sprintf("【%s】销售出库", order.OrderNumber),
			})

			stock.Amount -= finishedStock.Amount
			err = db.Select("amount").Updates(&stock).Error
			if err != nil {
				return err
			}

			finishedStock.Amount = 0
		} else {
			_, err = SaveFinishedConsume(db, &models.FinishedConsume{
				BaseModel: models.BaseModel{
					Operator: order.Operator,
				},
				OrderId:          &order.ID,
				FinishedId:       stock.FinishedId,
				ProductionId:     stock.ProductionId,
				StockNum:         0 - stock.Amount,
				OperationType:    false,
				OperationDetails: fmt.Sprintf("【%s】销售出库", order.OrderNumber),
			})

			finishedStock.Amount -= stock.Amount

			// 删除库存
			err = db.Delete(&stock).Error
			if err != nil {
				return err
			}
		}
	}
	return err
}

// DeductFinishedStockByProduct 产品扣除库存, 并且新增消耗表
func DeductFinishedStockByProduct(db *gorm.DB, product *models.Product,
	finishedStock *models.FinishedStock) error {

	var err error
	for {
		if finishedStock.Amount == 0 {
			break
		}

		stock := &models.FinishedStock{}
		err = db.Model(&models.FinishedStock{}).
			Where("finished_id = ?", finishedStock.FinishedId).
			Where("amount > ?", 0).
			Order("add_time asc").First(&stock).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(fmt.Sprintf("id: %d 成品库存不足", finishedStock.FinishedId))
		}
		if err != nil {
			return err
		}

		if stock.Amount > finishedStock.Amount {
			// 更新库存
			_, err = SaveFinishedConsume(db, &models.FinishedConsume{
				BaseModel: models.BaseModel{
					Operator: product.Operator,
				},
				FinishedId:       stock.FinishedId,
				ProductionId:     stock.ProductionId,
				StockNum:         0 - finishedStock.Amount,
				OperationType:    false,
				OperationDetails: fmt.Sprintf("产品【%s】新增库存", product.Name),
			})

			stock.Amount -= finishedStock.Amount
			err = db.Select("amount").Updates(&stock).Error
			if err != nil {
				return err
			}

			finishedStock.Amount = 0
		} else {
			_, err = SaveFinishedConsume(db, &models.FinishedConsume{
				BaseModel: models.BaseModel{
					Operator: product.Operator,
				},
				FinishedId:       stock.FinishedId,
				ProductionId:     stock.ProductionId,
				StockNum:         0 - stock.Amount,
				OperationType:    false,
				OperationDetails: fmt.Sprintf("产品【%s】新增库存", product.Name),
			})

			finishedStock.Amount -= stock.Amount

			// 删除库存
			err = db.Delete(&stock).Error
			if err != nil {
				return err
			}
		}
	}
	return err
}
