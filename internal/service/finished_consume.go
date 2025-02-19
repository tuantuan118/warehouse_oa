package service

import (
	"errors"
	"gorm.io/gorm"
	"warehouse_oa/internal/models"
)

// SaveConsumeByProduction 通过报工表来保存消耗表
func SaveConsumeByProduction(db *gorm.DB, production *models.FinishedProduction) error {
	if production.FinishedId == 0 {
		return errors.New("成品ID错误")
	}
	if production.ID == 0 {
		return errors.New("成品报工ID错误")
	}
	if production.ActualAmount == 0 {
		return errors.New("成品数量错误")
	}

	_, err := SaveFinishedConsume(db, &models.FinishedConsume{
		BaseModel: models.BaseModel{
			Operator: production.Operator,
		},
		OrderId:          nil,
		FinishedId:       production.FinishedId,
		ProductionId:     production.ID,
		StockNum:         float64(production.ActualAmount),
		OperationType:    true,
		OperationDetails: "生产完工",
	})

	return err
}

// SaveFinishedConsume 保存成品消耗表
func SaveFinishedConsume(db *gorm.DB, consume *models.FinishedConsume) (*models.FinishedConsume, error) {
	var err error
	if consume.OrderId != nil {
		_, err = GetOrderById(*consume.OrderId)
		if err != nil {
			return nil, err
		}
	}

	_, err = GetFinishedById(consume.FinishedId)
	if err != nil {
		return nil, err
	}

	_, err = GetProductionById(consume.ProductionId)
	if err != nil {
		return nil, err
	}

	err = db.Model(&models.FinishedConsume{}).Create(&consume).Error

	return consume, err
}

func SaveFinishedConsumeByOrder(db *gorm.DB, consume *models.FinishedConsume) {

}
