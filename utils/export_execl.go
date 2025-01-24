package utils

import (
	"fmt"
	"github.com/xuri/excelize/v2"
)

func ExportExcel(key []string, value []map[string]interface{}) (*excelize.File, error) {
	f := excelize.NewFile()

	// 设置表名
	sheetName := "Sheet1"
	_, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	for i, k := range key {
		cell := fmt.Sprintf("%s1", getExcelColumnName(i))
		err = f.SetCellValue(sheetName, cell, k)
		if err != nil {
			return nil, err
		}
	}

	for n, v := range value {
		for i, k := range key {
			cell := fmt.Sprintf("%s%d", getExcelColumnName(i), n+2)
			err = f.SetCellValue(sheetName, cell, v[k])
			if err != nil {
				return nil, err
			}
		}
	}

	return f, nil
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
