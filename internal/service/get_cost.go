package service

import (
	"github.com/sirupsen/logrus"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

// GetCostByConsume 配料出入库列表成本查询（根据ID查询）消耗表Id - InBoundId - UnitPrice
func GetCostByConsume(consume models.IngredientConsume) (float64, error) {
	var price float64
	err := global.Db.Select("unit_price").
		Model(&models.IngredientInBound{}).
		Where("id = ?", *consume.InBoundId).Find(&price).Error

	cost := price * (-consume.StockNum)
	return cost, err
}

// GetCostByProduction 成品报功接口成本查询（根据成品报功ID查询）成本报功ID - ProductionId - InBoundId - UnitPrice
func GetCostByProduction(id int) (float64, error) {
	var consume []models.IngredientConsume
	err := global.Db.Model(&models.IngredientConsume{}).
		Where("production_id = ?", id).Find(&consume).Error

	var cost float64
	for _, c := range consume {
		consumeCost, err := GetCostByConsume(c)
		if err != nil {
			return 0, err
		}
		cost += consumeCost
	}
	logrus.Infoln("GetCostByProduction-cost", cost)

	return cost, err
}

// GetCostByOrder 订单成本接口成本查询 （根据订单ID查询）订单ID - ProductionId - InBoundId - UnitPrice
func GetCostByOrder(order *models.Order) (float64, error) {
	var consume []*models.FinishedConsume
	err := global.Db.Model(&models.FinishedConsume{}).
		Where("order_id = ?", order.ID).Find(&consume).Error

	var cost float64
	for _, c := range consume {
		production, err := GetProductionById(c.ProductionId)
		if err != nil {
			return 0, err
		}
		unitPrice := production.Cost / float64(production.ActualAmount)
		cost += unitPrice * c.StockNum
	}
	logrus.Infoln("GetCostByOrder-cost", cost)
	cost = -(cost)
	ingredientCost, err := GetCostByOrderIngredient(order.ID)
	if err != nil {
		return 0, err
	}

	logrus.Infoln("GetCostByOrder-cost", cost)
	logrus.Infoln("GetCostByOrder-ingredientCost", ingredientCost)

	cost = cost + -ingredientCost
	return cost, err
}

// GetCostByOrderIngredient 订单附加材料成本查询 （根据订单ID查询）订单ID - InBoundId
func GetCostByOrderIngredient(orderId int) (float64, error) {
	var consume []models.IngredientConsume
	err := global.Db.Model(&models.IngredientConsume{}).
		Where("order_id = ?", orderId).Find(&consume).Error

	var cost float64
	for _, c := range consume {
		consumeCost, err := GetCostByConsume(c)
		if err != nil {
			return 0, err
		}
		cost += consumeCost
	}
	logrus.Infoln("GetCostByOrderIngredient-cost", cost)

	return cost, err
}
