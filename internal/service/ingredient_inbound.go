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
func GetInBoundList(name, stockUnit, supplier, begTime, endTime string,
	pn, pSize, isPackage int) (interface{}, error) {
	db := global.Db.Model(&models.IngredientInBound{})
	totalDb := global.Db.Model(&models.IngredientInBound{})

	db = db.Where("is_package = ? ", isPackage)
	totalDb = totalDb.Where("is_package = ? ", isPackage)

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
	if supplier != "" {
		slice := strings.Split(supplier, ";")
		db = db.Where("supplier in ?", slice)
		totalDb = totalDb.Where("supplier in ?", slice)
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

	// 未结金额
	unFinishPrice := totalPrice - finishPrice

	db = db.Preload("Ingredient")

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	if pn != 0 && pSize != 0 {
		offset := (pn - 1) * pSize
		db = db.Order("stock_time, id desc").Limit(pSize).Offset(offset)
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

// GetCountInBoundByIngredientId 根据配料ID查询入库信息
func GetCountInBoundByIngredientId(ingredientId int) (int64, error) {
	var total int64
	db := global.Db.Model(&models.IngredientInBound{})

	err := db.Where("ingredient_id = ?", ingredientId).Count(&total).Error
	if err != nil {
		return total, err
	}

	return total, err
}

// SaveInBound 保存
func SaveInBound(inBound *models.IngredientInBound) (*models.IngredientInBound, error) {
	// 获取配料ID
	logrus.Infoln("inbound:", *inBound.IngredientId)
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

	// 添加配料库存
	err = SaveStockByInBound(tx, inBound)
	if err != nil {
		return nil, err
	}

	// 添加配料消耗表
	err = SaveConsumeByInBound(tx, inBound, "配料入库")

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

	var total int64
	if global.Db.Model(&models.IngredientConsume{}).
		Where("in_bound_id = ? and operation_type = ?", data.ID, false).Count(&total); total != 0 {
		return errors.New("配料已使用，无法删除")
	}

	err = tx.Where("in_bound_id = ?", data.ID).Delete(&models.IngredientConsume{}).Error
	if err != nil {
		return err
	}

	err = tx.Where("in_bound_id = ?", data.ID).Delete(&models.IngredientStock{}).Error
	if err != nil {
		return err
	}

	err = tx.Delete(&data).Error
	if err != nil {
		return err
	}

	return err
}

// ExportIngredients 配料入库页面导出
func ExportIngredients(name, stockUser, begTime, endTime string) (*excelize.File, error) {
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
		"采购金额（元）",
		"已结金额（元）",
		"未结金额（元）",
		"付款金额",
		"付款日期",
		"入库数量",
		"入库人员",
		"入库时间",
		"备注",
	}

	valueList := make([]map[string]interface{}, 0)
	for _, v := range data {

		m := make([]map[string]string, 0)
		fpl := strings.Split(v.PaymentHistory, ";")
		for _, f := range fpl {
			fp := strings.Split(f, "&")
			if len(fp) != 2 {
				continue
			}
			m = append(m, map[string]string{
				"time":  fp[0],
				"price": fp[1],
			})
		}

		var paymentPrice, paymentTime string
		if len(m) > 0 {
			paymentPrice = m[0]["price"]
			paymentTime = m[0]["time"]
		}
		valueList = append(valueList, map[string]interface{}{
			"配料名称":    v.Ingredient.Name,
			"配料供应商":   v.Supplier,
			"配料规格":    v.Specification,
			"单价（元）":   v.UnitPrice,
			"采购金额（元）": fmt.Sprintf("%0.2f", v.TotalPrice),
			"已结金额（元）": fmt.Sprintf("%0.2f", v.FinishPrice),
			"未结金额（元）": fmt.Sprintf("%0.2f", v.TotalPrice-v.FinishPrice),
			"付款金额":    paymentPrice,
			"付款日期":    paymentTime,
			"入库数量":    fmt.Sprintf("%.2f%s", v.StockNum, returnUnit(v.StockUnit)),
			"入库人员":    v.StockUser,
			"入库时间":    v.StockTime,
			"备注":      v.Remark,
		})
		for i := 1; i < len(m); i++ {
			valueList = append(valueList, map[string]interface{}{
				"付款金额": m[i]["price"],
				"付款日期": m[i]["time"],
			})
		}
	}
	valueList = append(valueList, map[string]interface{}{
		"采购金额（元）": fmt.Sprintf("采购金额合计（元）: %0.2f", totalPrice),
		"已结金额（元）": fmt.Sprintf("已结金额合计（元）: %0.2f", finishPrice),
		"未结金额（元）": fmt.Sprintf("未结金额合计（元）: %0.2f", unFinishPrice),
	})

	return utils.ExportExcel(keyList, valueList, []string{"D", "E", "F", "G", "H", "I"})
}

// FinishInBound 结帐
func FinishInBound(bound []models.FinishInBound, operator string) (err error) {
	data := new(models.IngredientInBound)
	tx := global.Db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	for _, ifb := range bound {
		if ifb.ID == 0 {
			return errors.New("id is 0")
		}
		data, err = GetInBoundById(ifb.ID)
		if err != nil {
			return err
		}

		if data.Status == 1 {
			return errors.New("配料已结清，无法付款")
		}

		data.FinishPrice += ifb.TotalPrice

		str := fmt.Sprintf("%s&%0.2f;", ifb.PaymentTime, ifb.TotalPrice)
		data.PaymentHistory += str

		if data.TotalPrice-data.FinishPrice > 0 {
			data.Status = 0
		} else {
			data.Status = 1
		}
		data.Operator = operator

		err = tx.Select("FinishPrice", "PaymentHistory",
			"Operator", "Status").Updates(&data).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// GetSupplier 获取所有供应商
func GetSupplier() ([]string, error) {
	supplierList := make([]string, 0)

	db := global.Db.Model(&models.IngredientInBound{})
	db = db.Distinct("supplier").Where("supplier != ''")

	if err := db.Find(&supplierList).Error; err != nil {
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
