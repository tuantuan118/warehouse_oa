package service

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	_ "image/png"
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
	db = db.Preload("Customer")
	db = db.Preload("OrderProduct.UserList")
	db = db.Preload("OrderProduct.Ingredient")
	db = db.Preload("OrderProduct.UseFinished")

	b, err := getAdmin(userId)
	if err != nil {
		return nil, err
	}
	if !b {
		db = db.Where("id in ("+
			"select order_id from tb_order_product where id in ("+
			"select order_product_id from tb_order_product_user where user_id = ?))", userId)
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

	return map[string]interface{}{
		"data":       data,
		"pageNo":     pn,
		"pageSize":   pSize,
		"totalCount": total,
	}, err
}

func GetListById(id int) (interface{}, error) {
	if id == 0 {
		return nil, errors.New("ID不能为0")
	}

	db := global.Db.Model(&models.Order{})

	data := &models.Order{}
	db = db.Preload("Customer")
	db = db.Preload("OrderProduct.UserList")
	db = db.Preload("OrderProduct.Ingredient")
	db = db.Preload("OrderProduct.UseFinished")

	err := db.Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("订单不存在")
	}

	if data.PaymentHistory != "" {
		data.PaymentHistoryList = make([]map[string]string, 0)
		fpl := strings.Split(data.PaymentHistory, ";")
		for _, f := range fpl {
			fp := strings.Split(f, "&")
			if len(fp) != 2 {
				continue
			}
			data.PaymentHistoryList = append(data.PaymentHistoryList, map[string]string{
				"time":  fp[0],
				"price": fp[1],
			})
		}
	}

	for _, op := range data.OrderProduct {
		op.ImageList = make([]string, 0)
		if op.Images != "" {
			op.ImageList = strings.Split(op.Images, ";")
		}
	}

	cost, err := GetCostByOrder(data)
	if err != nil {
		return nil, err
	}

	data.Cost = cost
	data.Profit = data.TotalPrice - data.Cost
	data.GrossMargin = data.Profit / data.TotalPrice
	data.UnFinishPrice = data.TotalPrice - data.FinishPrice

	return data, err
}

func GetOrderById(id int) (*models.Order, error) {
	db := global.Db.Model(&models.Order{})

	data := &models.Order{}
	db = db.Preload("OrderProduct.UserList")
	db = db.Preload("OrderProduct.Ingredient")
	db = db.Preload("OrderProduct.UseFinished")
	db = db.Preload("Customer")
	err := db.Where("id = ?", id).First(&data).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("user does not exist")
	}

	return data, err
}

func SaveOrder(order *models.Order) (*models.Order, error) {
	var err error

	_, err = GetCustomerById(order.CustomerId)
	if err != nil {
		return nil, err
	}

	var totalPrice float64
	for _, orderProduct := range order.OrderProduct {
		if orderProduct.UserList != nil || len(orderProduct.UserList) > 0 {
			for _, v := range orderProduct.UserList {
				_, err = GetUserById(v.ID)
				if err != nil {
					return nil, err
				}
			}
		}

		for _, ingredient := range orderProduct.Ingredient {
			_, err = GetIngredientsById(*ingredient.IngredientId)
			if err != nil {
				return nil, err
			}
		}

		product, err := GetProductById(orderProduct.ProductId)
		if err != nil {
			return nil, err
		}
		orderProduct.ProductName = product.Name
		orderProduct.Specification = product.Specification

		orderProduct.Images = strings.Join(orderProduct.ImageList, ";")
		orderProduct.UseFinished = make([]models.UseFinished, 0)
		for _, p := range product.ProductContent {
			orderProduct.UseFinished = append(orderProduct.UseFinished, models.UseFinished{
				FinishedId: p.FinishedId,
				Quantity:   p.Quantity,
			})
		}
		totalPrice += orderProduct.Price * float64(orderProduct.Amount)
	}

	today := time.Now().Format("20060102")
	total, err := getTodayOrderCount()
	if err != nil {
		return nil, err
	}

	order.OrderNumber = fmt.Sprintf("QY%s%d", today, total+10001)
	order.TotalPrice = totalPrice
	order.FinishPrice = 0
	order.Status = 1

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

	_, err = GetCustomerById(order.CustomerId)
	if err != nil {
		return nil, err
	}

	order.OrderProduct = nil
	order.OrderNumber = oldData.OrderNumber
	order.TotalPrice = oldData.TotalPrice
	order.FinishPrice = oldData.FinishPrice
	order.Status = oldData.Status

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
	for _, od := range order.OrderProduct {
		for _, u := range od.UseFinished {
			err = DeductFinishedStock(tx, order, &models.FinishedStock{
				FinishedId: u.FinishedId,
				Amount:     u.Quantity * float64(od.Amount),
			})
			if err != nil {
				return err
			}
		}
		// 消耗附加材料
		for _, ingredient := range od.Ingredient {
			err = DeductOrderAttach(tx, order,
				&models.IngredientStock{
					IngredientId: ingredient.IngredientId,
					StockNum:     ingredient.Quantity * float64(od.Amount),
					StockUnit:    ingredient.StockUnit,
				})
			if err != nil {
				return err
			}
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

	str := fmt.Sprintf("%s&%0.2f;", paymentTime, totalPrice)
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

	var i, sumAmount int
	var sumTotalPrice float64
	for _, p := range order.OrderProduct {
		var productId int
		product, err := GetProductByIndex(p.ProductName, p.Specification)
		if err != nil {
			return nil, err
		}
		if product != nil {
			productId = product.ID
		}

		num := 10 + i
		err = f.DuplicateRow("Sheet1", num)
		if err != nil {
			return nil, err
		}
		if err := f.SetCellValue("Sheet1", fmt.Sprintf("B%d", num), productId); err != nil {
			return nil, err
		}
		if err := f.SetCellValue("Sheet1", fmt.Sprintf("C%d", num), p.ProductName); err != nil {
			return nil, err
		}
		if err := f.SetCellValue("Sheet1", fmt.Sprintf("D%d", num), p.Specification); err != nil {
			return nil, err
		}
		if err := f.SetCellValue("Sheet1", fmt.Sprintf("E%d", num), p.Amount); err != nil {
			return nil, err
		}
		F10 := fmt.Sprintf("¥%0.2f", p.Price)
		if err := f.SetCellValue("Sheet1", fmt.Sprintf("F%d", num), F10); err != nil {
			return nil, err
		}
		totalPrice := fmt.Sprintf("¥%0.2f", p.Price*float64(p.Amount))
		if err := f.SetCellValue("Sheet1", fmt.Sprintf("G%d", num), totalPrice); err != nil {
			return nil, err
		}

		sumAmount += p.Amount
		sumTotalPrice += p.Price * float64(p.Amount)
		i = i + 1
	}

	sumTotalPriceStr := utils.AmountConvert(sumTotalPrice, true)
	B12 := fmt.Sprintf("合计(大写): %s", sumTotalPriceStr)
	if err := f.SetCellValue("Sheet1", fmt.Sprintf("B%d", 12+i), B12); err != nil {
		return nil, err
	}

	if err := f.SetCellValue("Sheet1", fmt.Sprintf("D%d", 12+i), "总数量："); err != nil {
		return nil, err
	}
	if err := f.SetCellValue("Sheet1", fmt.Sprintf("E%d", 12+i), sumAmount); err != nil {
		return nil, err
	}

	if err := f.SetCellValue("Sheet1", fmt.Sprintf("F%d", 12+i), "总额小写："); err != nil {
		return nil, err
	}
	G12 := fmt.Sprintf("¥%0.2f", sumTotalPrice)
	if err := f.SetCellValue("Sheet1", fmt.Sprintf("G%d", 12+i), G12); err != nil {
		return nil, err
	}

	B14 := fmt.Sprintf("收货欠款人签名：")
	if err := f.SetCellValue("Sheet1", fmt.Sprintf("B%d", 14+i), B14); err != nil {
		return nil, err
	}
	D14 := fmt.Sprintf("运输：")
	if err := f.SetCellValue("Sheet1", fmt.Sprintf("D%d", 14+i), D14); err != nil {
		return nil, err
	}
	F14 := fmt.Sprintf("制单人：%s", order.Salesman)
	if err := f.SetCellValue("Sheet1", fmt.Sprintf("F%d", 14+i), F14); err != nil {
		return nil, err
	}
	B16 := fmt.Sprintf("备注:以上货品的数量请当即核对，规格、数量、质量、验收如有不符请当天内通逾期恕不负责，不便之处，敬请原谅。")
	if err := f.SetCellValue("Sheet1", fmt.Sprintf("B%d", 16+i), B16); err != nil {
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
		db = db.Where("id in ("+
			"select order_id from tb_order_product where id in ("+
			"select order_product_id from tb_order_product_user where user_id = ?))", userId)
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

func ExportOrderExecl(order *models.Order, customerStr, begTime, endTime string, pn, pSize int,
	costStatus, userId int) (
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

	var (
		totalPrice  float64
		finishPrice float64
		totalCost   float64
		sumCost     float64
	)

	keyList := []string{
		"订单编号",    //"订单编号"
		"客户名称",    //"客户名称"
		"产品名称",    //"产品名称"
		"产品规格",    //"产品规格"
		"单价（元）",   //"单价（元）"
		"数量",      //"数量"
		"订单分配",    //"订单分配"
		"销售金额（元）", //"销售金额（元）"
		"成本（元）",   //"成本（元）"
		"利润（元）",   //"利润（元）"
		"毛利率",     //"毛利率"
		"订单总额（元）", //"订单总额（元）"
		"已结金额（元）", //"已结金额（元）"
		"未结金额（元）", //"未结金额（元）"
		"销售日期",    //"销售日期"
		"订单状态",    //"订单状态"
		"销售人员",    //"销售人员"
		"备注",      //"备注"
		"更新人员",    //"更新人员"
		"更新时间",    //"更新时间"
	}

	var row int = 1 // 行数
	sheetName := "Sheet1"

	f := excelize.NewFile()
	_, err = f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	for i, k := range keyList {
		cell := fmt.Sprintf("%s%d", getExcelColumnName(i), row)
		err = f.SetCellValue(sheetName, cell, k)
		if err != nil {
			return nil, err
		}
	}

	for _, v := range orderList {
		rowCopy := row + 1
		var cell string

		cost, err := GetCostByOrder(&v)
		if err != nil {
			return nil, err
		}
		v.Cost = cost
		v.Profit = v.TotalPrice - v.Cost
		v.GrossMargin = v.Profit / v.TotalPrice * 100

		sumCost += v.Cost
		totalPrice += v.TotalPrice
		finishPrice += v.FinishPrice
		totalCost += v.Cost
		for _, op := range v.OrderProduct {
			row++
			var userListStr string
			for _, name := range op.UserList {
				userListStr += name.Nickname + ";"
			}
			cell = fmt.Sprintf("%s%d", "A", row)
			err = f.SetCellValue(sheetName, cell, v.OrderNumber)
			if err != nil {
				return nil, err
			}
			cell = fmt.Sprintf("%s%d", "B", row)
			err = f.SetCellValue(sheetName, cell, v.Customer.Name)
			if err != nil {
				return nil, err
			}
			cell = fmt.Sprintf("%s%d", "C", row)
			err = f.SetCellValue(sheetName, cell, op.ProductName)
			if err != nil {
				return nil, err
			}
			cell = fmt.Sprintf("%s%d", "D", row)
			err = f.SetCellValue(sheetName, cell, op.Specification)
			if err != nil {
				return nil, err
			}
			cell = fmt.Sprintf("%s%d", "E", row)
			err = f.SetCellValue(sheetName, cell, fmt.Sprintf("%.2f", op.Price))
			if err != nil {
				return nil, err
			}
			cell = fmt.Sprintf("%s%d", "F", row)
			err = f.SetCellValue(sheetName, cell, op.Amount)
			if err != nil {
				return nil, err
			}
			cell = fmt.Sprintf("%s%d", "G", row)
			err = f.SetCellValue(sheetName, cell, userListStr)
			if err != nil {
				return nil, err
			}
			cell = fmt.Sprintf("%s%d", "H", row)
			err = f.SetCellValue(sheetName, cell, fmt.Sprintf("%.2f", op.Price*float64(op.Amount)))
			if err != nil {
				return nil, err
			}
			cell = fmt.Sprintf("%s%d", "I", row)
			err = f.SetCellValue(sheetName, cell, fmt.Sprintf("%0.2f", v.Cost))
			if err != nil {
				return nil, err
			}
			cell = fmt.Sprintf("%s%d", "J", row)
			err = f.SetCellValue(sheetName, cell, fmt.Sprintf("%0.2f", v.Profit))
			if err != nil {
				return nil, err
			}
			cell = fmt.Sprintf("%s%d", "K", row)
			err = f.SetCellValue(sheetName, cell, fmt.Sprintf("%0.2f", v.GrossMargin))
			if err != nil {
				return nil, err
			}
			cell = fmt.Sprintf("%s%d", "L", row)
			err = f.SetCellValue(sheetName, cell, fmt.Sprintf("%0.2f", v.TotalPrice))
			if err != nil {
				return nil, err
			}
			cell = fmt.Sprintf("%s%d", "M", row)
			err = f.SetCellValue(sheetName, cell, fmt.Sprintf("%0.2f", v.FinishPrice))
			if err != nil {
				return nil, err
			}
			cell = fmt.Sprintf("%s%d", "N", row)
			err = f.SetCellValue(sheetName, cell, fmt.Sprintf("%0.2f", v.TotalPrice-v.FinishPrice))
			if err != nil {
				return nil, err
			}
			cell = fmt.Sprintf("%s%d", "O", row)
			err = f.SetCellValue(sheetName, cell, v.SaleDate.Format("2006-01-02 15:04:05"))
			if err != nil {
				return nil, err
			}
			cell = fmt.Sprintf("%s%d", "P", row)
			err = f.SetCellValue(sheetName, cell, returnStatus(v.Status))
			if err != nil {
				return nil, err
			}
			cell = fmt.Sprintf("%s%d", "Q", row)
			err = f.SetCellValue(sheetName, cell, v.Salesman)
			if err != nil {
				return nil, err
			}
			cell = fmt.Sprintf("%s%d", "R", row)
			err = f.SetCellValue(sheetName, cell, v.Remark)
			if err != nil {
				return nil, err
			}
			cell = fmt.Sprintf("%s%d", "S", row)
			err = f.SetCellValue(sheetName, cell, v.Operator)
			if err != nil {
				return nil, err
			}
			cell = fmt.Sprintf("%s%d", "T", row)
			err = f.SetCellValue(sheetName, cell, v.UpdatedAt.Format("2006-01-02 15:04:05"))
			if err != nil {
				return nil, err
			}
		}

		// 合并同一订单
		columnList := []string{"A", "B", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T"}
		for _, c := range columnList {
			err = f.MergeCell(sheetName,
				fmt.Sprintf("%s%d", c, rowCopy),
				fmt.Sprintf("%s%d", c, row))
			if err != nil {
				return nil, err
			}
		}
	}

	// 修改样式
	allStyle, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	err = f.SetCellStyle(sheetName,
		"A1",
		fmt.Sprintf("T%d", row+1),
		allStyle)
	if err != nil {
		return nil, err
	}

	redFontList := []string{"A", "L", "M", "N", "O"}
	redFontStyle, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Font: &excelize.Font{
			Color: "FF0000",
		},
	})
	for _, r := range redFontList {
		err = f.SetCellStyle(sheetName,
			fmt.Sprintf("%s1", r),
			fmt.Sprintf("%s%d", r, row+1),
			redFontStyle)
		if err != nil {
			return nil, err
		}
	}
	yellowBackList := []string{"I", "J", "K"}
	yellowBackStyle, err := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FFFF00"}, // 黄色背景
			Pattern: 1,                   // 实心填充
		},
	})
	for _, r := range yellowBackList {
		err = f.SetCellStyle(sheetName,
			fmt.Sprintf("%s2", r),
			fmt.Sprintf("%s%d", r, row),
			yellowBackStyle)
		if err != nil {
			return nil, err
		}
	}

	row++
	cell := fmt.Sprintf("%s%d", "A", row)
	err = f.SetCellValue(sheetName, cell, fmt.Sprintf("总订单数: %d", len(orderList)))
	if err != nil {
		return nil, err
	}
	cell = fmt.Sprintf("%s%d", "M", row)
	err = f.SetCellValue(sheetName, cell, fmt.Sprintf("已结合计: %0.2f", finishPrice))
	if err != nil {
		return nil, err
	}
	cell = fmt.Sprintf("%s%d", "N", row)
	err = f.SetCellValue(sheetName, cell, fmt.Sprintf("未结合计: %0.2f", totalPrice-finishPrice))
	if err != nil {
		return nil, err
	}
	cell = fmt.Sprintf("%s%d", "I", row)
	err = f.SetCellValue(sheetName, cell, fmt.Sprintf("成本合计: %0.2f", sumCost))
	if err != nil {
		return nil, err
	}
	cell = fmt.Sprintf("%s%d", "J", row)
	err = f.SetCellValue(sheetName, cell, fmt.Sprintf("利润合计: %0.2f", totalPrice-sumCost))
	if err != nil {
		return nil, err
	}
	cell = fmt.Sprintf("%s%d", "L", row)
	err = f.SetCellValue(sheetName, cell, fmt.Sprintf("订单总额合计: %0.2f", totalPrice))
	if err != nil {
		return nil, err
	}

	err = f.DuplicateRowTo(sheetName, row, 1)
	if err != nil {
		return nil, err
	}

	if costStatus == 0 {
		err = f.RemoveCol("Sheet1", "K")
		if err != nil {
			return nil, err
		}
		err = f.RemoveCol("Sheet1", "J")
		if err != nil {
			return nil, err
		}
		err = f.RemoveCol("Sheet1", "I")
		if err != nil {
			return nil, err
		}
	}

	return f, nil
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

func getExcelColumnName(n int) string {
	n += 1
	result := ""
	for n > 0 {
		n-- // Excel 列从 1 开始，这里减一进行调整
		result = string(rune('A'+(n%26))) + result
		n /= 26
	}
	return result
}
