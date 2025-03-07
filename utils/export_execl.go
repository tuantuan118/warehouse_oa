package utils

import (
	"fmt"
	"github.com/xuri/excelize/v2"
)

func ExportExcel(key []string, value []map[string]interface{}, redCol []string) (*excelize.File, error) {
	f := excelize.NewFile()

	// 设置表名
	sheetName := "Sheet1"
	_, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}

	colWidthMap := make(map[string]int)
	for i, k := range key {
		col := getExcelColumnName(i)
		cell := fmt.Sprintf("%s1", col)
		err = f.SetCellValue(sheetName, cell, k)
		if err != nil {
			return nil, err
		}
		setColWidthMap(colWidthMap, col, len(fmt.Sprintf("%v", k)))
	}

	var num int = 1
	for n, v := range value {
		for i, k := range key {
			col := getExcelColumnName(i)
			cell := fmt.Sprintf("%s%d", col, n+2)
			err = f.SetCellValue(sheetName, cell, v[k])
			if err != nil {
				return nil, err
			}
			setColWidthMap(colWidthMap, col, len(fmt.Sprintf("%v", v[k])))
		}
		num++
	}

	err = f.DuplicateRowTo(sheetName, num, 1)
	if err != nil {
		return nil, err
	}

	if len(redCol) != 0 {
		redStyle, err := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{
				Color: "FF0000",
			},
		})
		if err != nil {
			return nil, err
		}
		for _, s := range redCol {
			err = f.SetCellStyle(sheetName,
				fmt.Sprintf("%s1", s),
				fmt.Sprintf("%s%d", s, num+1),
				redStyle)
			if err != nil {
				return nil, err
			}
		}
	}

	// 应用列宽设置
	for col, width := range colWidthMap {
		if float64(width)+0.5 <= 10 {
			err = f.SetColWidth(sheetName, col, col, 10)
		} else if float64(width)+0.5 >= 20 {
			err = f.SetColWidth(sheetName, col, col, float64(width)*0.8)
		} else {
			err = f.SetColWidth(sheetName, col, col, float64(width)+0.5)
		}
		if err != nil {
			return nil, err
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

func setColWidthMap(m map[string]int, col string, l int) {
	if n, ok := m[col]; ok {
		if l > n {
			m[col] = l
		}
	} else {
		m[col] = l
	}
}
