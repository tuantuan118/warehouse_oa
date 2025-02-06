package models

type Order struct {
	BaseModel
	ProductId       int                 `gorm:"type:int(11);not null" json:"productId"`
	OrderNumber     string              `gorm:"type:varchar(256);not null" json:"orderNumber"`
	Name            string              `gorm:"type:varchar(256);not null" json:"name"`
	Specification   string              `gorm:"type:varchar(256)" json:"specification"`
	Price           float64             `gorm:"type:decimal(10,2)" json:"price"`
	Amount          int                 `gorm:"type:int(11);not null" json:"amount"`
	Cost            float64             `gorm:"type:decimal(12,2)" json:"cost"` // 成本
	TotalPrice      float64             `gorm:"type:decimal(10,2)" json:"totalPrice"`
	FinishPrice     float64             `gorm:"type:decimal(10,2)" json:"finishPrice"`
	FinishPriceStr  string              `gorm:"type:varchar(1024)" json:"finishPriceStr"`
	UnFinishPrice   float64             `gorm:"type:decimal(10,2)" json:"unFinishPrice"`
	Status          int                 `gorm:"type:int(11);not null" json:"status"` // 1:待出库 2:未完成支付 3:已支付 4:作废
	CustomerId      int                 `gorm:"type:int(11);not null" json:"customerId"`
	Customer        *Customer           `gorm:"foreignKey:CustomerId" json:"customer"`
	Salesman        string              `gorm:"type:varchar(256)" json:"salesman"`
	Images          string              `gorm:"type:text" json:"images"`
	UserList        []User              `gorm:"many2many:order_user;" json:"userList"`
	Ingredient      []AddIngredient     `gorm:"foreignKey:OrderID;references:ID" json:"ingredient"`
	ImageList       []string            `gorm:"-" json:"imageList"`
	FinishPriceList []map[string]string `gorm:"-" json:"finishPriceList"`
	Profit          float64             `gorm:"-" json:"profit"`
	GrossMargin     float64             `gorm:"-" json:"grossMargin"`
}

type AddIngredient struct {
	OrderID             int                  `gorm:"primaryKey;index" json:"orderID"`
	IngredientID        int                  `gorm:"primaryKey;" json:"ingredient_id"`
	IngredientInventory *IngredientInventory `gorm:"foreignKey:IngredientID" json:"ingredient_inventory"`
	Quantity            int                  `gorm:"type:int(11);not null" json:"quantity"` // 用量
}
