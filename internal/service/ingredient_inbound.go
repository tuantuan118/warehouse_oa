package service

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"math/big"
	"strings"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
	"warehouse_oa/utils"
)

// GetInBoundList 返回入库列表查询数据
func GetInBoundList(name, stockUnit, begTime, endTime string,
	pn, pSize int) (interface{}, error) {
	db := global.Db.Model(&models.IngredientInBound{})
	totalDb := global.Db.Model(&models.IngredientInBound{})

	if name != "" {
		idList, err := GetIngredientsByName(name)
		if err != nil {
			return nil, err
		}
		db = db.Where("ingredient_id in ?", idList)
		totalDb = totalDb.Where("ingredient_id in ?", idList)
	}
	if stockUnit != "" {
		db = db.Where("stock_unit = ?", stockUnit)
		totalDb = totalDb.Where("stock_unit = ?", stockUnit)
	}
	if begTime != "" && endTime != "" {
		db = db.Where("DATE_FORMAT(stock_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
		totalDb = totalDb.Where("DATE_FORMAT(stock_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
	}

	// 应结金额
	var totalPrice float64
	if err := totalDb.Select("COALESCE(SUM(total_price), 0)").Scan(&totalPrice).Error; err != nil {
		return nil, err
	}
	// 已结金额
	var finishPrice float64
	if err := totalDb.Select("COALESCE(SUM(finish_price), 0)").Scan(&finishPrice).Error; err != nil {
		return nil, err
	}
	var cost float64
	if err := totalDb.Select("SUM(cost)").Scan(&cost).Error; err != nil {
		return nil, err
	}
	// 未结金额
	unFinishPrice := totalPrice - finishPrice

	db = db.Preload("Ingredient")

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	if pn != 0 && pSize != 0 {
		offset := (pn - 1) * pSize
		db = db.Order("id desc").Limit(pSize).Offset(offset)
	}

	data := make([]models.IngredientInBound, 0)
	err := db.Find(&data).Error
	if err != nil {
		return nil, err
	}

	var inBoundDataList = make([]models.GetInBoundList, 0)
	for _, d := range data {
		var inBoundData = models.GetInBoundList{
			ID:              d.ID,
			Operator:        d.Operator,
			Remark:          d.Remark,
			CreatedAt:       d.CreatedAt,
			UpdatedAt:       d.UpdatedAt,
			IngredientId:    d.IngredientId,
			Ingredient:      d.Ingredient,
			Supplier:        d.Supplier,
			Specification:   d.Specification,
			UnitPrice:       d.UnitPrice,
			TotalPrice:      d.TotalPrice,
			FinishPrice:     d.FinishPrice,
			UnFinishPrice:   d.TotalPrice - d.FinishPrice,
			PaymentHistory:  d.PaymentHistory,
			Status:          d.Status,
			StockNum:        d.StockNum,
			StockUnit:       d.StockUnit,
			StockUser:       d.StockUser,
			StockTime:       d.StockTime,
			FinishPriceList: make([]map[string]string, 0),
		}

		if d.PaymentHistory == "" {
			continue
		}

		fpl := strings.Split(d.PaymentHistory, ";")
		for _, f := range fpl {
			fp := strings.Split(f, "&")
			if len(fp) != 2 {
				continue
			}
			inBoundData.FinishPriceList = append(inBoundData.FinishPriceList, map[string]string{
				"time":  fp[0],
				"price": fp[1],
			})
		}
		inBoundDataList = append(inBoundDataList, inBoundData)
	}

	return map[string]interface{}{
		"data":             inBoundDataList,
		"pageNo":           pn,
		"pageSize":         pSize,
		"totalCount":       total,
		"sumTotalPrice":    totalPrice,
		"sumUnFinishPrice": unFinishPrice,
		"sumFinishPrice":   finishPrice,
		"cost":             cost,
	}, err

}

// GetInBoundById 根据ID查询入库信息
func GetInBoundById(id int) (*models.IngredientInBound, error) {
	db := global.Db.Model(&models.IngredientInBound{})

	data := &models.IngredientInBound{}
	err := db.Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("配料不存在")
	}

	return data, err
}

// SaveInBound 保存
func SaveInBound(inBound *models.IngredientInBound) (*models.IngredientInBound, error) {
	// 获取配料ID
	ingredients, err := GetIngredientsById(*inBound.IngredientId)
	if err != nil {
		return nil, err
	}

	totalPrice := big.NewFloat(inBound.TotalPrice)
	stockNum := big.NewFloat(inBound.StockNum)
	price := new(big.Float).Quo(totalPrice, stockNum)

	inBound.Ingredient = ingredients
	inBound.UnitPrice, _ = price.Float64()
	inBound.FinishPrice = 0.0

	db := global.Db
	tx := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	err = tx.Model(&models.IngredientInBound{}).Create(&inBound).Error
	if err != nil {
		return nil, err
	}

	err = SaveStockByInBound(tx, inBound)
	if err != nil {
		return nil, err
	}

	err = SaveConsumeByInBound(tx, inBound)

	return inBound, err
}

// UpdateInBound 更新
func UpdateInBound(inBound *models.IngredientInBound) (*models.IngredientInBound, error) {
	if inBound.ID == 0 {
		return nil, errors.New("id is 0")
	}
	var err error
	oldData := new(models.IngredientInBound)
	oldData, err = GetInBoundById(inBound.ID)
	if err != nil {
		return nil, err
	}

	// 修改关联的配料ID
	if oldData.IngredientId != inBound.IngredientId {
		ingredients := new(models.Ingredients)
		ingredients, err = GetIngredientsById(*inBound.IngredientId)
		if err != nil {
			return nil, err
		}

		inBound.Ingredient = ingredients
	}

	// 修改单价
	totalPrice := big.NewFloat(inBound.TotalPrice)
	stockNum := big.NewFloat(inBound.StockNum)
	price := new(big.Float).Quo(totalPrice, stockNum)
	inBound.UnitPrice, _ = price.Float64()

	db := global.Db
	tx := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	err = tx.Updates(&inBound).Error
	if err != nil {
		return nil, err
	}

	inBound.StockNum = inBound.StockNum - oldData.StockNum

	// 从库存中 扣除旧数据的数量，新增新数据的数量
	err = DeductStockByInBound(tx, oldData)
	if err != nil {
		if err.Error() == "配料不足" {
			return nil, errors.New("当前配料的库存已被使用，修改可能影响成本计算")
		}
		return nil, err
	}

	err = SaveStockByInBound(tx, inBound)
	return inBound, err
}

// DelInBound 删除
func DelInBound(id int, username string) error {
	if id == 0 {
		return errors.New("id is 0")
	}

	data, err := GetInBoundById(id)
	if err != nil {
		return err
	}
	if data == nil {
		return errors.New("配料ID错误")
	}

	data.Operator = username

	db := global.Db
	tx := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	err = tx.Delete(&data).Error
	if err != nil {
		return err
	}

	// 扣除库存
	err = DeductStockByInBound(tx, data)

	// 删除出入库信息
	err = DelConsumeByInBound(tx, data.ID)

	return err
}

// ExportIngredients 配料入库页面导出
func ExportIngredients(name, supplier, stockUser, begTime, endTime string) (*excelize.File, error) {
	db := global.Db.Model(&models.IngredientInBound{})
	totalDb := global.Db.Model(&models.IngredientInBound{})

	if name != "" {
		idList, err := GetIngredientsByName(name)
		if err != nil {
			return nil, err
		}
		db = db.Where("ingredient_id in ?", idList)
		totalDb = totalDb.Where("ingredient_id in ?", idList)
	}
	if supplier != "" {
		db = db.Where("supplier = ?", supplier)
		totalDb = totalDb.Where("supplier = ?", supplier)
	}
	if stockUser != "" {
		db = db.Where("stock_user = ?", stockUser)
		totalDb = totalDb.Where("stock_user = ?", stockUser)
	}
	if begTime != "" && endTime != "" {
		db = db.Where("DATE_FORMAT(stock_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
		totalDb = totalDb.Where("DATE_FORMAT(stock_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
	}

	var totalPrice float64
	err := totalDb.Select("COALESCE(SUM(total_price), 0)").Scan(&totalPrice).Error
	if err != nil {
		return nil, err
	}
	var finishPrice float64
	if err = totalDb.Select("COALESCE(SUM(finish_price), 0)").Scan(&finishPrice).Error; err != nil {
		return nil, err
	}
	var consumeCost float64
	if err = totalDb.Select("SUM(cost)").Scan(&consumeCost).Error; err != nil {
		return nil, err
	}
	unFinishPrice := totalPrice - finishPrice

	data := make([]models.IngredientInBound, 0)
	err = db.Preload("Ingredient").Find(&data).Error
	if err != nil {
		logrus.Infoln("导出订单错误: ", err.Error())
	}

	keyList := []string{
		"配料名称",
		"配料供应商",
		"配料规格",
		"单价（元）",
		"金额（元）",
		"已结金额",
		"未结金额",
		"入库数量",
		"入库人员",
		"入库时间",
		"备注",
	}

	valueList := make([]map[string]interface{}, 0)
	for _, v := range data {
		valueList = append(valueList, map[string]interface{}{
			"配料名称":  v.Ingredient.Name,
			"配料供应商": v.Supplier,
			"配料规格":  v.Specification,
			"单价（元）": v.UnitPrice,
			"金额（元）": v.TotalPrice,
			"已结金额":  v.FinishPrice,
			"未结金额":  v.TotalPrice - v.FinishPrice,
			"入库数量":  fmt.Sprintf("%f%s", v.StockNum, returnUnit(v.StockUnit)),
			"入库人员":  v.StockUser,
			"入库时间":  v.StockTime,
			"备注":    v.Remark,
		})
	}
	valueList = append(valueList, map[string]interface{}{
		"金额（元）": totalPrice,
		"已结金额":  finishPrice,
		"未结金额":  unFinishPrice,
		"成本":    consumeCost,
	})

	return utils.ExportExcel(keyList, valueList)
}

// FinishInBound 结帐
func FinishInBound(id int, totalPrice float64, paymentTime, operator string) (*models.IngredientInBound, error) {
	if id == 0 {
		return nil, errors.New("id is 0")
	}
	data, err := GetInBoundById(id)
	if err != nil {
		return nil, err
	}

	if data.Status == 1 {
		return nil, errors.New("配料已结清，无法付款")
	}

	data.FinishPrice += totalPrice

	str := fmt.Sprintf("%s&%f;", paymentTime, totalPrice)
	data.PaymentHistory += str

	if data.TotalPrice-data.FinishPrice > 0 {
		data.Status = 0
	} else {
		data.Status = 1
	}
	data.Operator = operator

	return data, global.Db.Select("FinishPrice", "FinishPriceStr",
		"Operator", "Status").Updates(&data).Error
}

// GetSupplier 获取所有供应商
func GetSupplier() ([]string, error) {
	supplierList := make([]string, 0)

	db := global.Db.Model(&models.IngredientInBound{})
	db = db.Distinct("supplier").Where("supplier != ''")

	if err := db.Scan(&supplierList).Error; err != nil {
		return nil, err
	}

	return supplierList, nil
}

// returnUnit 单位影射表
func returnUnit(i int) string {
	switch i {
	case 1:
		return "斤"
	case 2:
		return "克"
	case 3:
		return "件"
	case 4:
		return "个"
	case 5:
		return "张"
	case 6:
		return "盆"
	case 7:
		return "桶"
	case 8:
		return "包"
	case 9:
		return "箱"
	}
	return ""
}
