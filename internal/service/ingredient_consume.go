package service

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
	"warehouse_oa/utils"
)

// GetConsumeList 返回出入库列表查询数据
func GetConsumeList(ids, stockUnit, begTime, endTime string,
	inOrOut, pn, pSize, isPackage int) (interface{}, error) {

	db := global.Db.Model(&models.IngredientConsume{})
	totalDb := global.Db.Model(&models.IngredientConsume{})

	db = db.Where("is_package = ?", isPackage)
	totalDb = db.Where("is_package = ?", isPackage)

	if ids != "" {
		idList := strings.Split(ids, ";")
		db = db.Where("ingredient_id in ?", idList)
		totalDb = totalDb.Where("ingredient_id in ?", idList)
	}
	if stockUnit != "" {
		db = db.Where("stock_unit = ?", stockUnit)
		totalDb = totalDb.Where("stock_unit = ?", stockUnit)
	}
	if begTime != "" && endTime != "" {
		db = db.Where("DATE_FORMAT(add_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
		totalDb = totalDb.Where("DATE_FORMAT(add_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
	}
	if inOrOut == 1 {
		db = db.Where("stock_num > 0")
		totalDb = totalDb.Where("stock_num > 0")
	}
	if inOrOut == 2 {
		db = db.Where("stock_num < 0")
		totalDb = totalDb.Where("stock_num < 0")
	}

	consumeCost, err := GetConsumeAllCost()
	db = db.Preload("Ingredient")

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	if pn != 0 && pSize != 0 {
		offset := (pn - 1) * pSize
		db = db.Order("id desc").Limit(pSize).Offset(offset)
	}

	data := make([]models.IngredientConsume, 0)
	err = db.Find(&data).Error
	if err != nil {
		return nil, err
	}

	for n := range data {
		if data[n].StockNum > 0 {
			data[n].Cost = 0
			continue
		}
		cost, err := GetCostByConsume(data[n])
		if err != nil {
			return nil, err
		}
		logrus.Info("consume cost:", cost)
		data[n].Cost = cost
	}

	return map[string]interface{}{
		"data":       data,
		"pageNo":     pn,
		"pageSize":   pSize,
		"totalCount": total,
		"cost":       consumeCost,
	}, err
}

// GetIngredientSum 返回出入库列表查询数据
func GetIngredientSum(ids, stockUnit, begTime, endTime string,
	inOrOut int) (interface{}, error) {

	enterDb := global.Db.Model(&models.IngredientConsume{})
	outDb := global.Db.Model(&models.IngredientConsume{})

	if ids != "" {
		idList := strings.Split(ids, ";")
		enterDb = enterDb.Where("ingredient_id in ?", idList)
		outDb = outDb.Where("ingredient_id in ?", idList)
	}
	if stockUnit != "" {
		enterDb = enterDb.Where("stock_unit = ?", stockUnit)
		outDb = outDb.Where("stock_unit = ?", stockUnit)
	}
	if begTime != "" && endTime != "" {
		enterDb = enterDb.Where("DATE_FORMAT(add_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
		outDb = outDb.Where("DATE_FORMAT(add_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
	}

	var enterNum, outNum float64
	err := enterDb.Where("stock_num >= 0").Select("IFNULL(SUM(stock_num), 0) AS stock_num").First(&enterNum).Error
	if err != nil {
		return nil, err
	}
	err = outDb.Where("stock_num <= 0").Select("IFNULL(SUM(stock_num), 0) AS stock_num").First(&outNum).Error
	if err != nil {
		return nil, err
	}

	if inOrOut == 1 {
		outNum = 0
	}
	if inOrOut == 2 {
		enterNum = 0
	}

	return map[string]float64{
		"enter": enterNum,
		"out":   outNum,
	}, nil
}

// GetConsumeChart 返回出入库列表图表
func GetConsumeChart(ids, stockUnit, begTime, endTime string) ([]map[string]interface{}, error) {
	db := global.Db.Model(&models.IngredientConsume{})
	totalDb := global.Db.Model(&models.IngredientConsume{})

	if ids != "" {
		idList := strings.Split(ids, ";")
		db = db.Where("ingredient_id in ?", idList)
		totalDb = totalDb.Where("ingredient_id in ?", idList)
	}
	if stockUnit != "" {
		db = db.Where("stock_unit = ?", stockUnit)
		totalDb = totalDb.Where("stock_unit = ?", stockUnit)
	}
	if begTime != "" && endTime != "" {
		db = db.Where("DATE_FORMAT(add_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
		totalDb = totalDb.Where("DATE_FORMAT(add_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
	}

	data := make([]map[string]interface{}, 0)
	err := db.Select("ingredient_id, stock_unit, stock_num").Find(&data).Error
	if err != nil {
		return nil, err
	}

	return data, err
}

// SaveConsumeByInBound 通过入库表来保存消耗表
func SaveConsumeByInBound(db *gorm.DB, inBound *models.IngredientInBound, details string) error {
	if inBound.IngredientId == nil && *inBound.IngredientId == 0 {
		return errors.New("配料ID错误")
	}
	if inBound.ID == 0 {
		return errors.New("配料入库ID错误")
	}
	if inBound.StockUnit == 0 {
		return errors.New("配料单位错误")
	}

	var b bool
	if details == "配料入库" {
		b = true
	}

	_, err := SaveConsume(db, &models.IngredientConsume{
		BaseModel: models.BaseModel{
			Operator: inBound.Operator,
		},
		IngredientId:     inBound.IngredientId,
		InBoundId:        &inBound.ID,
		StockNum:         inBound.StockNum,
		StockUnit:        inBound.StockUnit,
		OperationType:    &b,
		OperationDetails: details,
		IsPackage:        inBound.IsPackage,
	})

	return err
}

// GetConsumeByProduction 根据报工ID查找消耗表
func GetConsumeByProduction(id int) ([]*models.IngredientConsume, error) {
	var dataList []*models.IngredientConsume

	err := global.Db.Model(&models.IngredientConsume{}).
		Where("production_id = ?", id).
		Find(&dataList).Error

	return dataList, err
}

// SaveConsume 保存消耗表
func SaveConsume(db *gorm.DB, consume *models.IngredientConsume) (*models.IngredientConsume, error) {
	_, err := GetIngredientsById(*consume.IngredientId)
	if err != nil {
		return nil, err
	}

	err = db.Model(&models.IngredientConsume{}).Create(&consume).Error

	return consume, err
}

// DelConsumeByInBound 通过入库表来删除消耗表
func DelConsumeByInBound(db *gorm.DB, id int) error {
	var total int64
	if err := global.Db.Model(&models.IngredientConsume{}).
		Where("in_bound_id = ? and operation_type = ?", id, false).Count(&total).Error; err != nil {
		return errors.New("配料已使用，无法删除")
	}

	err := db.Where("in_bound_id = ? and operation_type = ?", id, true).
		Delete(&models.IngredientConsume{}).Error

	return err
}

// GetConsumeAllCost 获取全部消耗成本
func GetConsumeAllCost() (string, error) {
	var cost string
	err := global.Db.Raw(`SELECT
		sum(tb_ingredient_in_bound.unit_price * tb_ingredient_consume.stock_num) AS cost
		FROM
		tb_ingredient_consume
		JOIN
		tb_ingredient_in_bound ON tb_ingredient_in_bound.id = tb_ingredient_consume.in_bound_id 
		WHERE operation_type = FALSE;`).First(&cost).Error

	return cost, err
}

// ExportConsume 配料出入库页面导出
func ExportConsume(ids, stockUnit, begTime, endTime string) (*excelize.File, error) {
	db := global.Db.Model(&models.IngredientConsume{})
	totalDb := global.Db.Model(&models.IngredientConsume{})

	if ids != "" {
		idList := strings.Split(ids, ";")
		db = db.Where("ingredient_id in ?", idList)
		totalDb = totalDb.Where("ingredient_id in ?", idList)
	}
	if stockUnit != "" {
		db = db.Where("stock_unit = ?", stockUnit)
		totalDb = totalDb.Where("stock_unit = ?", stockUnit)
	}
	if begTime != "" && endTime != "" {
		db = db.Where("DATE_FORMAT(add_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
		totalDb = totalDb.Where("DATE_FORMAT(add_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
	}

	costStr, err := GetConsumeAllCost()
	db = db.Preload("Ingredient")

	data := make([]models.IngredientConsume, 0)
	db = db.Preload("Ingredient")
	err = db.Preload("InBound").Order("id desc").Find(&data).Error
	if err != nil {
		return nil, err
	}

	valueList := make([]map[string]interface{}, 0)
	for _, v := range data {
		var operationType string
		if v.StockNum <= 0 {
			consumeCost, err := GetCostByConsume(v)
			if err != nil {
				return nil, err
			}
			v.Cost = consumeCost
			operationType = "出库"
		} else {
			operationType = "入库"
		}

		valueList = append(valueList, map[string]interface{}{
			"配料名称":    v.Ingredient.Name,
			"操作类型":    operationType,
			"操作数量":    fmt.Sprintf("%0.2f(%s)", v.StockNum, returnUnit(v.InBound.StockUnit)),
			"操作明细":    v.OperationDetails,
			"成本金额（元）": fmt.Sprintf("%0.2f", v.Cost),
			"操作时间":    v.CreatedAt.Format("2006-01-02 15:04:05"),
			"操作人员":    v.Operator,
		})
	}

	keyList := []string{
		"配料名称",
		"操作类型",
		"操作数量",
		"操作明细",
		"成本金额（元）",
		"操作时间",
		"操作人员",
	}

	// string 转 float64
	var cost float64
	if len(costStr) == 0 {
		cost = 0
	} else {
		cost, err = strconv.ParseFloat(costStr, 64)
		if err != nil {
			return nil, err
		}
	}
	valueList = append(valueList, map[string]interface{}{
		"成本金额（元）": fmt.Sprintf("成本金额合计（元）: %.2f", cost),
	})

	return utils.ExportExcel(keyList, valueList, []string{"E"})
}
