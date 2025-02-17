package models

type Order struct {
	BaseModel
	OrderNumber    string           `gorm:"type:varchar(256);not null" json:"orderNumber"`
	ProductInfo    string           `gorm:"type:text" json:"productInfo"`
	Cost           float64          `gorm:"type:decimal(12,2)" json:"cost"` // 成本
	TotalPrice     float64          `gorm:"type:decimal(10,2)" json:"totalPrice"`
	FinishPrice    float64          `gorm:"type:decimal(10,2)" json:"finishPrice"`
	FinishPriceStr string           `gorm:"type:varchar(1024)" json:"finishPriceStr"`
	Status         int              `gorm:"type:int(11);not null" json:"status"` // 1:待出库 2:未完成支付 3:已支付 4:作废
	CustomerId     int              `gorm:"type:int(11);not null" json:"customerId"`
	Customer       *Customer        `gorm:"foreignKey:CustomerId" json:"customer"`
	Salesman       string           `gorm:"type:varchar(256)" json:"salesman"`
	Images         string           `gorm:"type:text" json:"images"`
	UserList       []User           `gorm:"many2many:order_user;" json:"userList"`
	Ingredient     []AddIngredient  `gorm:"foreignKey:OrderID;references:ID" json:"ingredient"`
	Content        []ProductContent `gorm:"foreignKey:orderId;references:ID" json:"material"`
}

type AddIngredient struct {
	OrderID      int          `gorm:"primaryKey;index" json:"orderID"`
	IngredientId *int         `gorm:"type:int(11)" json:"ingredientId"`
	Ingredient   *Ingredients `gorm:"foreignKey:IngredientId" json:"ingredient"`
	StockUnit    int          `gorm:"type:int(2)" json:"stockUnit"`
	Quantity     float64      `gorm:"type:decimal(10,4);not null" json:"quantity"` // 用量
}

type OrderContent struct {
	OrderId     int       `gorm:"primaryKey;index" json:"orderId"`
	ProductName string    `gorm:"type:varchar(256);not null" json:"productName"`
	FinishedId  int       `gorm:"type:int(11)" json:"finishedId"`
	Finished    *Finished `gorm:"foreignKey:FinishedId;" json:"finished"`
	Quantity    float64   `gorm:"type:decimal(10,4);not null" json:"quantity"` // 用量
}

//ImageList       []string            `gorm:"-" json:"imageList"`
//FinishPriceList []map[string]string `gorm:"-" json:"finishPriceList"`
//Profit          float64             `gorm:"-" json:"profit"`
//GrossMargin     float64             `gorm:"-" json:"grossMargin"`
