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

	logrus.Infoln("consume.InBoundId:", *consume.InBoundId)
	logrus.Infoln("price:", price)
	logrus.Infoln("consume.StockNum:", consume.StockNum)
	cost := price * (-consume.StockNum)
	return cost, err
}

// GetCostByProduction 成品报功接口成本查询（根据成品报功ID查询）成本报功ID - ProductionId - InBoundId - UnitPrice
func GetCostByProduction(id int, num float64) (float64, error) {
	var consume []models.IngredientConsume
	err := global.Db.Model(&models.IngredientConsume{}).
		Where("production_id = ?", id).Find(&consume).Error

	var cost float64
	for _, c := range consume {
		if num != 0 {
			c.StockNum = num
		}
		consumeCost, err := GetCostByConsume(c)
		if err != nil {
			return 0, err
		}
		logrus.Infoln("GetCostByProduction-consumeCost", consumeCost)
		cost += consumeCost
	}
	logrus.Infoln("GetCostByProduction-cost", cost)

	return cost, err
}

// GetCostByOrder 订单成本接口成本查询 （根据订单ID查询）订单ID - ProductionId - InBoundId - UnitPrice
func GetCostByOrder(order *models.Order) (float64, error) {
	var consume []models.FinishedConsume
	err := global.Db.Model(&models.FinishedConsume{}).
		Where("order_id = ?", order.ID).Find(&consume).Error

	logrus.Infoln("GetCostByOrder", consume)
	logrus.Infoln(len(consume))

	var cost float64
	for _, c := range consume {
		productionCost, err := GetCostByProduction(c.ProductionId, c.StockNum)
		if err != nil {
			return 0, err
		}
		logrus.Infoln("GetCostByOrder-productionCost", productionCost)
		cost += productionCost
	}
	logrus.Infoln("GetCostByOrder-cost", cost)

	return cost, err
}

//	订单附加材料成本查询 （根据订单ID查询）订单ID - InBoundId
