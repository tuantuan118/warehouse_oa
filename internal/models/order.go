package models

type Order struct {
	BaseModel
	OrderNumber    string          `gorm:"type:varchar(256);not null" json:"orderNumber"`
	TotalPrice     float64         `gorm:"type:decimal(10,2)" json:"totalPrice"`     // 总价
	FinishPrice    float64         `gorm:"type:decimal(10,2)" json:"finishPrice"`    // 已结金额
	PaymentHistory string          `gorm:"type:varchar(1024)" json:"paymentHistory"` // 结账记录
	Status         int             `gorm:"type:int(11);not null" json:"status"`      // 1:待出库 2:未完成支付 3:已支付 4:作废
	CustomerId     int             `gorm:"type:int(11);not null" json:"customerId"`  // 客户ID
	Customer       *Customer       `gorm:"foreignKey:CustomerId" json:"customer"`
	Salesman       string          `gorm:"type:varchar(256)" json:"salesman"`     // 订单分配
	Images         string          `gorm:"type:text" json:"images"`               // 图片列表
	UserList       []User          `gorm:"many2many:order_user;" json:"userList"` // 销售人员
	Ingredient     []AddIngredient `gorm:"foreignKey:OrderID;references:ID" json:"ingredient"`
	Product        []OrderProduct  `gorm:"foreignKey:OrderId;references:ID" json:"content"`

	// 请求参数
	ImageList          []string            `gorm:"-" json:"imageList"`
	PaymentHistoryList []map[string]string `gorm:"-" json:"paymentHistoryList"`
	Profit             float64             `gorm:"-" json:"profit"`
	GrossMargin        float64             `gorm:"-" json:"grossMargin"`
}

type AddIngredient struct {
	OrderID      int          `gorm:"primaryKey;index" json:"orderID"`
	IngredientId *int         `gorm:"type:int(11)" json:"ingredientId"`
	Ingredient   *Ingredients `gorm:"foreignKey:IngredientId" json:"ingredient"`
	StockUnit    int          `gorm:"type:int(2)" json:"stockUnit"`
	Quantity     float64      `gorm:"type:decimal(10,4);not null" json:"quantity"` // 用量
}

type OrderProduct struct {
	ID          int        `gorm:"primaryKey" json:"id"`
	OrderId     int        `gorm:"index" json:"orderId"`
	ProductName string     `gorm:"type:varchar(256);not null" json:"productName"`
	Quantity    float64    `gorm:"type:decimal(10,4);not null" json:"quantity"`       // 用量
	Price       float64    `gorm:"type:decimal(10,2)" json:"price"`                   // 单价
	Amount      int        `gorm:"type:int(11);not null" json:"amount"`               // 数量
	Finish      []Finished `gorm:"many2many:order_product_finished;" json:"userList"` // 订单产品成品
}
