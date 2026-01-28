package model

type LineItem struct {
	ItemID   int64 `json:"item_id,string"`
	OrderID  int64 `gorm:"column:order_id" json:"order_id"`
	Quantity uint  `json:"quantity"`
	Price    uint  `json:"price"`
}
