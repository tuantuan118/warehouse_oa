package models

import "time"

type Finished struct {
	BaseModel
	Name             string          `gorm:"type:varchar(256);not null" json:"name"`
	Ratio            float64         `gorm:"type:decimal(10,2);not null" json:"ratio"`
	ExpectAmount     int             `gorm:"type:int(11);not null" json:"expectAmount"`
	ActualAmount     int             `gorm:"type:int(11);not null" json:"actualAmount"`
	Balance          int             `gorm:"type:int(11);not null" json:"balance"` // 余量
	Cost             float64         `gorm:"type:decimal(12,2)" json:"cost"`       // 成本
	Price            float64         `gorm:"type:decimal(12,2)" json:"price"`      // 单价
	Status           int             `gorm:"type:int(11);not null" json:"status"`
	FinishHour       int             `gorm:"-" json:"finishHour"`
	EstimatedTime    *time.Time      `gorm:"type:Time" json:"estimatedTime"`
	FinishTime       *time.Time      `gorm:"type:Time" json:"finishTime"`
	FinishedManageId int             `gorm:"type:int(11)" json:"finishedManageId"`
	FinishedManage   *FinishedManage `gorm:"foreignKey:FinishedManageId;" json:"finishedManage"`
	InAndOut         bool            `gorm:"type:tinyint(1)" json:"inAndOut"` // InAndOut True 入库 False 出库
	OperationType    string          `gorm:"type:varchar(256)" json:"operationType"`
	OperationDetails string          `gorm:"type:varchar(256)" json:"operationDetails"`
}
