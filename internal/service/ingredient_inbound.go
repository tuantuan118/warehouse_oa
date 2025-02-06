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

func GetInBoundList(name, supplier, stockUser, stockUnit, begTime, endTime string,
	pn, pSize int, b bool) (interface{}, error) {
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
		idList, err := GetIngredientsBySupplier(supplier)
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
	if stockUnit != "" {
		db = db.Where("stock_unit = ?", stockUnit)
		totalDb = totalDb.Where("stock_unit = ?", stockUnit)
	}
	if begTime != "" && endTime != "" {
		db = db.Where("DATE_FORMAT(stock_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
		totalDb = totalDb.Where("DATE_FORMAT(stock_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
	}

	var totalPrice float64
	if err := totalDb.Select("COALESCE(SUM(total_price), 0)").Scan(&totalPrice).Error; err != nil {
		return nil, err
	}
	var unFinishPrice float64
	if err := totalDb.Select("COALESCE(SUM(un_finish_price), 0)").Scan(&unFinishPrice).Error; err != nil {
		return nil, err
	}
	var finishPrice float64
	if err := totalDb.Select("COALESCE(SUM(finish_price), 0)").Scan(&finishPrice).Error; err != nil {
		return nil, err
	}
	var consumeCost float64
	if err := totalDb.Select("SUM(cost)").Scan(&consumeCost).Error; err != nil {
		return nil, err
	}

	if b {
		db = db.Where("in_and_out = ?", b)
	}
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

	for n := range data {
		if data[n].FinishPriceStr == "" {
			continue
		}
		data[n].FinishPriceList = make([]map[string]string, 0)
		fpl := strings.Split(data[n].FinishPriceStr, ";")
		for _, f := range fpl {
			fp := strings.Split(f, "&")
			if len(fp) != 2 {
				continue
			}
			data[n].FinishPriceList = append(data[n].FinishPriceList, map[string]string{
				"time":  fp[0],
				"price": fp[1],
			})
		}
	}

	return map[string]interface{}{
		"data":             data,
		"pageNo":           pn,
		"pageSize":         pSize,
		"totalCount":       total,
		"sumTotalPrice":    totalPrice,
		"sumUnFinishPrice": unFinishPrice,
		"sumFinishPrice":   finishPrice,
		"consumeCost":      consumeCost,
	}, err

}

func GetInBoundById(id int) (*models.IngredientInBound, error) {
	db := global.Db.Model(&models.IngredientInBound{})

	data := &models.IngredientInBound{}
	err := db.Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user does not exist")
	}

	return data, err
}

func SaveInBound(inBound *models.IngredientInBound) (*models.IngredientInBound, error) {
	ingredients, err := GetIngredientsById(*inBound.IngredientID)
	if err != nil {
		return nil, err
	}

	totalPrice := big.NewFloat(inBound.TotalPrice)
	stockNum := big.NewFloat(inBound.StockNum)
	price := new(big.Float).Quo(totalPrice, stockNum)
	inBound.Price, _ = price.Float64()
	inBound.Ingredient = ingredients
	inBound.InAndOut = true
	inBound.OperationType = "入库"
	inBound.Balance = inBound.StockNum
	inBound.OperationDetails = fmt.Sprintf("配料入库")
	inBound.UnFinishPrice, _ = totalPrice.Float64()

	db := global.Db
	tx := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	logrus.Infoln(inBound.Supplier)

	err = SaveInventoryByInBound(tx, inBound)
	if err != nil {
		return nil, err
	}
	err = tx.Model(&models.IngredientInBound{}).Create(&inBound).Error

	return inBound, err
}

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

	if oldData.IngredientID != inBound.IngredientID {
		ingredients := new(models.Ingredients)
		ingredients, err = GetIngredientsById(*inBound.IngredientID)
		if err != nil {
			return nil, err
		}

		inBound.Ingredient = ingredients
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

	inBound.TotalPrice = inBound.Price * inBound.StockNum
	inBound.UnFinishPrice = inBound.TotalPrice - oldData.FinishPrice
	err = tx.Updates(&inBound).Error
	if err != nil {
		return nil, err
	}

	inBound.StockNum = inBound.StockNum - oldData.StockNum

	return inBound, SaveInventoryByInBound(tx, inBound)
}

func DelInBound(id int, username string) error {
	if id == 0 {
		return errors.New("id is 0")
	}

	data, err := GetInBoundById(id)
	if err != nil {
		return err
	}
	if data == nil {
		return errors.New("user does not exist")
	}

	data.Operator = username
	data.IsDeleted = true

	db := global.Db
	tx := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	err = tx.Updates(&data).Error
	if err != nil {
		return err
	}
	err = tx.Delete(&data).Error
	if err != nil {
		return err
	}

	data.StockNum = 0 - data.StockNum
	return UpdateInventoryByInBound(tx, data)
}

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
		idList, err := GetIngredientsBySupplier(supplier)
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
	var unFinishPrice float64
	if err = totalDb.Select("COALESCE(SUM(un_finish_price), 0)").Scan(&unFinishPrice).Error; err != nil {
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
		"采购金额",
		"已结金额",
		"未结金额",
		"成本",
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
			"单价（元）": v.Price,
			"金额（元）": v.TotalPrice,
			"采购金额":  v.Price * v.StockNum,
			"已结金额":  v.FinishPrice,
			"未结金额":  v.UnFinishPrice,
			"成本":    v.Cost,
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

func FinishedSaveInBound(tx *gorm.DB, inBound *models.IngredientInBound) error {
	ingredients, err := GetIngredientsById(*inBound.IngredientID)
	if err != nil {
		return err
	}

	inBound.Ingredient = ingredients

	err = SaveInventoryByInBound(tx, inBound)
	if err != nil {
		return err
	}
	err = tx.Model(&models.IngredientInBound{}).Create(inBound).Error

	return nil
}

func UpdateInBoundBalance(tx *gorm.DB, inventory *models.IngredientInventory, pn int,
	amount float64) (float64, error) {

	var cost float64
	data := make([]models.IngredientInBound, 0)
	db := global.Db.Model(&models.IngredientInBound{})
	db = db.Where("ingredient_id = ?", inventory.IngredientID)
	db = db.Where("stock_unit = ?", inventory.StockUnit)
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
		d := &data[n]

		if amount >= d.Balance {
			cost = cost + d.Price*d.Balance
			amount = amount - d.Balance
			d.Balance = 0
		} else {
			cost = cost + d.Price*amount
			d.Balance = d.Balance - amount
			amount = 0
		}

		err = tx.Model(&models.IngredientInBound{}).
			Where("id = ?", d.ID).
			Update("balance", d.Balance).Error
		if err != nil {
			return 0, err
		}
		if amount == 0 {
			continue
		}
	}
	if amount > 0 {
		c, err := UpdateInBoundBalance(tx, inventory, pn+1, amount)
		if err != nil {
			return 0, err
		}
		cost = cost + c
	}

	return cost, nil
}

func GetOutInBoundList(id int, supplier, stockUser, begTime, endTime string,
	pn, pSize int) (interface{}, error) {

	var name string
	var stockUnit string
	if id != 0 {
		inventory, err := GetInventoryById(id)
		if err != nil {
			return nil, err
		}
		name = inventory.Ingredient.Name
		stockUnit = fmt.Sprintf("%d", inventory.StockUnit)
	}

	return GetInBoundList(
		name,
		supplier,
		stockUser,
		stockUnit,
		begTime,
		endTime,
		pn,
		pSize,
		false,
	)
}

func FinishInBound(id int, totalPrice float64, paymentTime, operator string) (*models.IngredientInBound, error) {
	if id == 0 {
		return nil, errors.New("id is 0")
	}
	data, err := GetInBoundById(id)
	if err != nil {
		return nil, err
	}

	if data.Status == 1 {
		return nil, errors.New("ingredient has been finished, can not update")
	}

	data.UnFinishPrice = data.UnFinishPrice - totalPrice
	data.FinishPrice += totalPrice

	str := fmt.Sprintf("%s&%f;", paymentTime, totalPrice)
	data.FinishPriceStr += str

	if data.UnFinishPrice > 0 {
		data.Status = 0
	} else {
		data.Status = 1
	}
	data.Operator = operator

	return data, global.Db.Select("UnFinishPrice",
		"FinishPrice", "FinishPriceStr", "Operator",
		"Status").Updates(&data).Error
}

func GetSupplier() ([]string, error) {
	supplierList := make([]string, 0)

	db := global.Db.Model(&models.IngredientInBound{})
	db = db.Distinct("supplier").Where("supplier != ''")

	if err := db.Scan(&supplierList).Error; err != nil {
		return nil, err
	}

	return supplierList, nil
}
