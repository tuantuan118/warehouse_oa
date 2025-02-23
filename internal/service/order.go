package service

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	"os"
	"os/exec"
	"strings"
	"time"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
	"warehouse_oa/utils"
)

func GetOrderList(order *models.Order, customerStr, begTime, endTime string, pn, pSize int, userId int) (
	interface{}, error) {

	db := global.Db.Model(&models.Order{})

	if order.OrderNumber != "" {
		slice := strings.Split(order.OrderNumber, ";")
		db = db.Where("order_number in ?", slice)
	}
	if order.ProductName != "" {
		slice := strings.Split(order.ProductName, ";")
		db = db.Where("product_name in ?", slice)
	}
	if order.Specification != "" {
		db = db.Where("specification = ?", order.Specification)
	}
	if order.Salesman != "" {
		slice := strings.Split(order.Salesman, ";")
		db = db.Where("salesman in ?", slice)
	}
	if customerStr != "" {
		slice := strings.Split(customerStr, ";")
		db = db.Where("customer_id in ?", slice)
	}
	if order.Status != 0 {
		db = db.Where("status = ?", order.Status)
	}
	if begTime != "" && endTime != "" {
		db = db.Where("DATE_FORMAT(add_time, '%Y-%m-%d') BETWEEN ? AND ?", begTime, endTime)
	}
	db = db.Preload("UserList")
	db = db.Preload("Customer")
	db = db.Preload("Ingredient.Ingredient")
	db = db.Preload("UseFinished")

	b, err := getAdmin(userId)
	if err != nil {
		return nil, err
	}
	if !b {
		db = db.Where(" id in (select order_id from tb_order_user where user_id = ?)", userId)
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	if pn != 0 && pSize != 0 {
		offset := (pn - 1) * pSize
		db = db.Order("id desc").Limit(pSize).Offset(offset)
	}

	data := make([]models.Order, 0)
	err = db.Find(&data).Error

	logrus.Infoln("len(data)", len(data))

	for n := range data {
		data[n].ImageList = make([]string, 0)
		if data[n].Images != "" {
			data[n].ImageList = strings.Split(data[n].Images, ";")
		}

		if data[n].PaymentHistory != "" {
			data[n].PaymentHistoryList = make([]map[string]string, 0)
			fpl := strings.Split(data[n].PaymentHistory, ";")
			for _, f := range fpl {
				fp := strings.Split(f, "&")
				if len(fp) != 2 {
					continue
				}
				data[n].PaymentHistoryList = append(data[n].PaymentHistoryList, map[string]string{
					"time":  fp[0],
					"price": fp[1],
				})
			}
		}

		cost, err := GetCostByOrder(&data[n])
		if err != nil {
			return nil, err
		}
		data[n].UnFinishPrice = data[n].TotalPrice - data[n].FinishPrice
		data[n].Cost = cost
		data[n].Profit = data[n].TotalPrice - data[n].Cost
		data[n].GrossMargin = data[n].Profit / data[n].TotalPrice
	}

	return map[string]interface{}{
		"data":       data,
		"pageNo":     pn,
		"pageSize":   pSize,
		"totalCount": total,
	}, err
}

func GetOrderById(id int) (*models.Order, error) {
	db := global.Db.Model(&models.Order{})

	data := &models.Order{}
	db = db.Preload("UserList")
	db = db.Preload("Customer")
	db = db.Preload("Ingredient.Ingredient")
	db = db.Preload("UseFinished")
	err := db.Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user does not exist")
	}

	return data, err
}

func SaveOrder(order *models.Order) (*models.Order, error) {
	var err error

	if order.UserList != nil || len(order.UserList) > 0 {
		for _, v := range order.UserList {
			_, err = GetUserById(v.ID)
			if err != nil {
				return nil, err
			}
		}
	}

	for _, ingredient := range order.Ingredient {
		_, err = GetIngredientsById(*ingredient.IngredientId)
		if err != nil {
			return nil, err
		}
	}

	_, err = GetCustomerById(order.CustomerId)
	if err != nil {
		return nil, err
	}

	product, err := GetProductById(order.ProductId)
	if err != nil {
		return nil, err
	}
	order.ProductName = product.Name

	today := time.Now().Format("20060102")
	total, err := getTodayOrderCount()
	if err != nil {
		return nil, err
	}

	order.Images = strings.Join(order.ImageList, ";")
	order.OrderNumber = fmt.Sprintf("QY%s%d", today, total+10001)
	order.TotalPrice = order.Price * float64(order.Amount)
	order.FinishPrice = 0
	order.Status = 1

	order.UseFinished = make([]models.UseFinished, 0)
	for _, p := range product.ProductContent {
		order.UseFinished = append(order.UseFinished, models.UseFinished{
			FinishedId: p.FinishedId,
			Quantity:   p.Quantity,
		})
	}

	err = global.Db.Model(&models.Order{}).Create(&order).Error
	if err != nil {
		return nil, err
	}

	return order, err
}

func UpdateOrder(order *models.Order) (*models.Order, error) {
	if order.ID == 0 {
		return nil, errors.New("id is 0")
	}
	oldData, err := GetOrderById(order.ID)
	if err != nil {
		return nil, err
	}

	if oldData.Status != 1 {
		return nil, errors.New("订单已出库，无法修改")
	}

	if order.UserList != nil || len(order.UserList) > 0 {
		for _, v := range order.UserList {
			_, err = GetUserById(v.ID)
			if err != nil {
				return nil, err
			}
		}
	}

	for _, ingredient := range order.Ingredient {
		_, err = GetIngredientsById(*ingredient.IngredientId)
		if err != nil {
			return nil, err
		}
	}

	_, err = GetCustomerById(order.CustomerId)
	if err != nil {
		return nil, err
	}

	product, err := GetProductByName(order.ProductName)
	if err != nil {
		return nil, err
	}

	order.Images = strings.Join(order.ImageList, ";")
	order.OrderNumber = oldData.OrderNumber
	order.TotalPrice = order.Price * float64(order.Amount)
	order.FinishPrice = oldData.FinishPrice
	order.Status = oldData.Status

	order.UseFinished = make([]models.UseFinished, 0)
	for _, p := range product.ProductContent {
		order.UseFinished = append(order.UseFinished, models.UseFinished{
			FinishedId: p.FinishedId,
			Quantity:   p.Quantity,
		})
	}

	// 清除 UserList 关联
	if err := global.Db.Model(&oldData).Association("UserList").Clear(); err != nil {
		return nil, err
	}
	// 清除 UserList 关联
	if err := global.Db.Model(&oldData).Association("Ingredient").Clear(); err != nil {
		return nil, err
	}
	// 清除 UserList 关联
	if err := global.Db.Model(&oldData).Association("UseFinished").Clear(); err != nil {
		return nil, err
	}

	return order, global.Db.Updates(&order).Error
}

// OutOfStock 出库
func OutOfStock(id int, username string) error {
	order, err := GetOrderById(id)

	db := global.Db
	tx := db.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// 修改出库状态
	order.Status = 2
	order.Operator = username

	// 消耗成品
	for _, u := range order.UseFinished {
		err = DeductFinishedStock(tx, order, &models.FinishedStock{
			FinishedId: u.FinishedId,
			Amount:     u.Quantity * float64(order.Amount),
		})
		if err != nil {
			return err
		}
	}

	// 消耗附加材料
	for _, ingredient := range order.Ingredient {
		err = DeductOrderAttach(tx, order,
			&models.IngredientStock{
				IngredientId: ingredient.IngredientId,
				StockNum:     ingredient.Quantity * float64(order.Amount),
				StockUnit:    ingredient.StockUnit,
			})
		if err != nil {
			return err
		}
	}

	err = tx.Select("status", "operator").Updates(&order).Error

	return err
}

// VoidOrder 作废
func VoidOrder(id int, username string) error {
	if id == 0 {
		return errors.New("id is 0")
	}

	data, err := GetOrderById(id)
	if err != nil {
		return err
	}
	if data == nil {
		return errors.New("user does not exist")
	}

	data.Operator = username
	data.Status = 4

	return global.Db.Updates(&data).Error
}

// CheckoutOrder 结帐
func CheckoutOrder(id int, totalPrice float64, paymentTime, operator string) (*models.Order, error) {
	if id == 0 {
		return nil, errors.New("id is 0")
	}
	data, err := GetOrderById(id)
	if err != nil {
		return nil, err
	}

	if data.Status != 2 {
		return nil, errors.New("订单状态错误，无法付款")
	}

	data.FinishPrice += totalPrice

	str := fmt.Sprintf("%s&%f;", paymentTime, totalPrice)
	data.PaymentHistory += str

	if data.TotalPrice-data.FinishPrice > 0 {
		data.Status = 2
	} else {
		data.Status = 3
	}
	data.Operator = operator

	return data, global.Db.Select(
		"finish_price",
		"payment_history",
		"operator",
		"status").Updates(&data).Error
}

func ExportOrder(order *models.Order) ([]byte, error) {
	db := global.Db.Model(&models.Order{})

	if order.ID != 0 {
		db = db.Where("id = ?", order.ID)
	}
	order, err := GetOrderById(order.ID)
	if err != nil {
		return nil, err
	}

	if p, err := GetProductByIndex(order.ProductName, order.Specification); err != nil {
		order.ProductId = 0
	} else {
		order.ProductId = p.ID
	}

	filePath := "./stencil.xlsx"
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	defer func(f *excelize.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)

	B5 := fmt.Sprintf("订单号：%s", order.OrderNumber)
	if err := f.SetCellValue("Sheet1", "B5", B5); err != nil {
		return nil, err
	}
	F5 := fmt.Sprintf("开单日期：%s", order.CreatedAt.Format("2006/01/02"))
	if err := f.SetCellValue("Sheet1", "F5", F5); err != nil {
		return nil, err
	}
	B6 := fmt.Sprintf("客户编号：%d", order.Customer.ID)
	if err := f.SetCellValue("Sheet1", "B6", B6); err != nil {
		return nil, err
	}
	D6 := fmt.Sprintf("客户名称：%s", order.Customer.Name)
	if err := f.SetCellValue("Sheet1", "D6", D6); err != nil {
		return nil, err
	}
	F6 := fmt.Sprintf("客户联系方式：%s", order.Customer.Phone)
	if err := f.SetCellValue("Sheet1", "F6", F6); err != nil {
		return nil, err
	}
	B7 := fmt.Sprintf("收货地址：%s", order.Customer.Address)
	if err := f.SetCellValue("Sheet1", "B7", B7); err != nil {
		return nil, err
	}
	if err := f.SetCellValue("Sheet1", "B10", order.ProductId); err != nil {
		return nil, err
	}
	if err := f.SetCellValue("Sheet1", "C10", order.ProductName); err != nil {
		return nil, err
	}
	if err := f.SetCellValue("Sheet1", "D10", order.Specification); err != nil {
		return nil, err
	}
	if err := f.SetCellValue("Sheet1", "E10", order.Amount); err != nil {
		return nil, err
	}
	F10 := fmt.Sprintf("¥%0.2f", order.Price)
	if err := f.SetCellValue("Sheet1", "F10", F10); err != nil {
		return nil, err
	}
	G10 := fmt.Sprintf("¥%0.2f", order.TotalPrice)
	if err := f.SetCellValue("Sheet1", "G10", G10); err != nil {
		return nil, err
	}
	totalPrice := utils.AmountConvert(order.TotalPrice, true)
	B12 := fmt.Sprintf("合计(大写): %s", totalPrice)
	if err := f.SetCellValue("Sheet1", "B12", B12); err != nil {
		return nil, err
	}
	if err := f.SetCellValue("Sheet1", "E12", order.Amount); err != nil {
		return nil, err
	}
	G12 := fmt.Sprintf("¥%0.2f", order.TotalPrice)
	if err := f.SetCellValue("Sheet1", "G12", G12); err != nil {
		return nil, err
	}
	F14 := fmt.Sprintf("制单人：%s", order.Salesman)
	if err := f.SetCellValue("Sheet1", "F14", F14); err != nil {
		return nil, err
	}

	newName := fmt.Sprintf("./cos/execl/%d.xlsx", order.ID)
	if err := f.SaveAs(newName); err != nil {
		return nil, err
	} else {
		logrus.Infoln("文件已成功另存为", newName)
	}

	cmd := exec.Command("libreoffice",
		"--invisible",
		"--convert-to",
		"pdf",
		"--outdir",
		"./cos/pdf/",
		newName,
	)
	err = cmd.Run()
	if err != nil {
		return nil, err
	}

	pdfName := fmt.Sprintf("./cos/pdf/%d.pdf", order.ID)
	pdfData, err := os.ReadFile(pdfName)
	if err != nil {
		return nil, err
	}

	return pdfData, nil
}

// GetOrderFieldList 获取字段列表
func GetOrderFieldList(field string, userId int) ([]string, error) {
	db := global.Db.Model(&models.Order{})
	switch field {
	case "productName":
		db = db.Distinct("product_name")
	case "orderNumber":
		db = db.Distinct("order_number")
	case "specification":
		db = db.Distinct("specification")
	case "salesman":
		db = db.Distinct("salesman")
	default:
		return nil, errors.New("field not exist")
	}
	b, err := getAdmin(userId)
	if err != nil {
		return nil, err
	}
	if !b {
		db = db.Where(" id in (select order_id from tb_order_user where user_id = ?)", userId)
	}
	fields := make([]string, 0)
	if err := db.Scan(&fields).Error; err != nil {
		return nil, err
	}

	return fields, nil
}

func getTodayOrderCount() (int64, error) {
	today := time.Now().Format("2006-01-02")
	startOfDay, _ := time.Parse("2006-01-02", today)

	var total int64
	err := global.Db.Model(&models.Order{}).Where(
		"add_time >= ?", startOfDay).Count(&total).Error

	return total, err
}

func GetOrderByCustomer(customerId int) error {
	db := global.Db.Model(&models.Order{})

	data := &models.Order{}
	err := db.Where("customer_id = ?", customerId).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("user does not exist")
	}

	return err
}

func ExportOrderExecl(order *models.Order, customerStr, begTime, endTime string, pn, pSize int, userId int) (
	*excelize.File, error) {

	i, err := GetOrderList(order, customerStr, begTime, endTime, pn, pSize, userId)
	if err != nil {
		return nil, err
	}

	m, b := i.(map[string]interface{})
	if !b {
		return nil, errors.New("导出错误")
	}

	orderList, b := m["data"].([]models.Order)
	if !b {
		return nil, errors.New("导出错误")
	}

	keyList := []string{
		"订单编号",
		"产品名称",
		"产品规格",
		"单价（元）",
		"数量",
		"订单金额",
		"已结金额",
		"未结金额",
		"成本",
		"利润",
		"毛利率",
		"订单状态",
		"客户名称",
		"订单分配",
		"销售人员",
		"备注",
		"更新人员",
		"更新时间",
	}

	var (
		totalPrice  float64
		finishPrice float64
		totalCost   float64
	)
	valueList := make([]map[string]interface{}, 0)
	for _, v := range orderList {
		var userStr string
		for _, u := range v.UserList {
			userStr += u.Nickname + ";"
		}

		valueList = append(valueList, map[string]interface{}{
			"订单编号":  v.OrderNumber,
			"产品名称":  v.ProductName,
			"产品规格":  v.Specification,
			"单价（元）": v.Price,
			"数量":    v.Amount,
			"订单金额":  v.TotalPrice,
			"已结金额":  v.FinishPrice,
			"未结金额":  v.UnFinishPrice,
			"成本":    v.Cost,
			"利润":    v.Profit,
			"毛利率":   v.GrossMargin,
			"订单状态":  fmt.Sprintf("%s", returnStatus(v.Status)),
			"客户名称":  v.Customer.Name,
			"订单分配":  userStr,
			"销售人员":  v.Salesman,
			"备注":    v.Remark,
			"更新人员":  v.Operator,
			"更新时间":  v.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
		totalPrice += v.TotalPrice
		finishPrice += v.FinishPrice
		totalCost += v.Cost
	}
	valueList = append(valueList, map[string]interface{}{
		"订单金额": totalPrice,
		"已结金额": finishPrice,
		"未结金额": totalPrice - finishPrice,
		"成本":   totalCost,
	})

	return utils.ExportExcel(keyList, valueList)
}

func returnStatus(i int) string {
	switch i {
	case 1:
		return "待出库"
	case 2:
		return "未完成支付"
	case 3:
		return "已支付"
	case 4:
		return "作废"
	}
	return ""
}
