package utils

import (
	"github.com/sirupsen/logrus"
	"github.com/xuri/excelize/v2"
	"mime/multipart"
)

func UploadXlsx(file *multipart.FileHeader) ([]map[string]string, error) {
	fileContent, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer func(fileContent multipart.File) {
		_ = fileContent.Close()
	}(fileContent)

	f, err := excelize.OpenReader(fileContent)
	if err != nil {
		return nil, err
	}
	defer func(ff *excelize.File) {
		_ = ff.Close()
	}(f)

	// 获取指定工作表的所有行
	sheetName := "Sheet1"
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}

	dataList := make([]map[string]string, 0)
	// 遍历行数据并打印
	for _, row := range rows {
		m := make(map[string]string)
		for j, cell := range row {
			m[rows[0][j]] = cell
		}
		dataList = append(dataList, m)
	}

	logrus.Infoln(dataList)

	return nil, nil
}
