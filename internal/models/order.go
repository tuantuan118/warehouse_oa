package models

import "time"

type Order struct {
	BaseModel
	OrderNumber    string          `gorm:"type:varchar(256);not null" json:"orderNumber"`
	TotalPrice     float64         `gorm:"type:decimal(10,2)" json:"totalPrice"`     // 总价
	FinishPrice    float64         `gorm:"type:decimal(10,2)" json:"finishPrice"`    // 已结金额
	PaymentHistory string          `gorm:"type:varchar(1024)" json:"paymentHistory"` // 结账记录
	Status         int             `gorm:"type:int(11);not null" json:"status"`      // 1:待出库 2:未完成支付 3:已支付 4:作废
	CustomerId     int             `gorm:"type:int(11);not null" json:"customerId"`  // 客户ID
	Customer       *Customer       `gorm:"foreignKey:CustomerId" json:"customer"`
	Salesman       string          `gorm:"type:varchar(256)" json:"salesman"` // 销售人员
	SaleDate       time.Time       `gorm:"type:Time;not null" json:"saleDate"`
	OrderProduct   []*OrderProduct `gorm:"foreignKey:OrderId;references:ID" json:"orderProduct"`

	PaymentHistoryList []map[string]string `gorm:"-" json:"paymentHistoryList"`
	Profit             float64             `gorm:"-" json:"profit"`
	GrossMargin        float64             `gorm:"-" json:"grossMargin"`
	Cost               float64             `gorm:"-" json:"cost"`
	UnFinishPrice      float64             `gorm:"-" json:"unFinishPrice"` // 已结金额
}

type OrderProduct struct {
	BaseModel
	OrderId       int             `gorm:"index" json:"orderId"`
	ProductId     int             `gorm:"type:int(11)" json:"productId"`
	ProductName   string          `gorm:"type:varchar(256);not null" json:"productName"`
	Specification string          `gorm:"type:varchar(256)" json:"specification"`
	Price         float64         `gorm:"type:decimal(10,2)" json:"price"`
	Amount        int             `gorm:"type:int(11);not null" json:"amount"`
	Images        string          `gorm:"type:text" json:"images"`                       // 图片列表
	UserList      []User          `gorm:"many2many:order_product_user;" json:"userList"` // 订单分配
	Ingredient    []AddIngredient `gorm:"foreignKey:OrderProductId;references:ID" json:"ingredient"`
	UseFinished   []UseFinished   `gorm:"foreignKey:OrderProductId;references:ID" json:"useFinished"` // 订单成品

	// 请求参数
	ImageList []string `gorm:"-" json:"imageList"`
}

type AddIngredient struct {
	OrderProductId int          `gorm:"index" json:"orderProductId"`
	IngredientId   *int         `gorm:"type:int(11)" json:"ingredientId"`
	Ingredient     *Ingredients `gorm:"foreignKey:IngredientId" json:"ingredient"`
	StockUnit      int          `gorm:"type:int(2)" json:"stockUnit"`
	Quantity       float64      `gorm:"type:decimal(10,4);not null" json:"quantity"` // 用量
}

type UseFinished struct {
	OrderProductId int     `gorm:"index" json:"orderProductId"`
	FinishedId     int     `gorm:"type:int(11)" json:"finishedId"`
	Quantity       float64 `gorm:"type:decimal(10,4);not null" json:"quantity"` // 用量
}
