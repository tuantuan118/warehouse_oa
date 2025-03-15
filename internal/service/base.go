package service

import (
	"errors"
	"gorm.io/gorm"
	"warehouse_oa/internal/global"
	"warehouse_oa/internal/models"
)

func Pagination(db *gorm.DB, model interface{}, pn, pSize int) (map[string]interface{}, error) {
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	if pn != 0 && pSize != 0 {
		offset := (pn - 1) * pSize
		db = db.Order("id desc").Limit(pSize).Offset(offset)
	}

	data := model
	err := db.Find(&data).Error

	return map[string]interface{}{
		"data":       data,
		"pageNo":     pn,
		"pageSize":   pSize,
		"totalCount": total,
	}, err
}

// GetUpdate 获取字段列表
func GetUpdate(update, updateTime string) (int64, error) {
	db := global.Db
	switch update {
	case "iib":
		db = global.Db.Model(&models.IngredientInBound{}) // 配料入库
	case "is":
		db = global.Db.Model(&models.IngredientStock{}) // 配料库存
	case "ic":
		db = global.Db.Model(&models.IngredientConsume{}) // 配料出入库
	case "fp":
		db = global.Db.Model(&models.FinishedProduction{}) // 成品报工
	case "fs":
		db = global.Db.Model(&models.FinishedStock{}) // 成品库存
	case "fc":
		db = global.Db.Model(&models.FinishedConsume{}) // 成品出入库
	case "o":
		db = global.Db.Model(&models.Order{}) // 订单
	default:
		return 0, errors.New("查询刷新接口参数错误")
	}
	var num int64
	if err := db.Where("update_time >= ?", updateTime).Count(&num).Error; err != nil {
		return 0, err
	}

	return num, nil
}
