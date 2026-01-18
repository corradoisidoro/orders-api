package model

import (
	"time"

	"github.com/google/uuid"
)

type Order struct {
	OrderID     uint64     `gorm:"primarykey;column:order_id" json:"order_id"`
	CustomerID  uuid.UUID  `json:"customer_id"`
	LineItems   []LineItem `gorm:"foreignKey:OrderID" json:"line_items"`
	CreatedAt   *time.Time `json:"created_at"`
	ShippedAt   *time.Time `json:"shipped_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

type LineItem struct {
	ItemID   uuid.UUID `gorm:"type:uuid;primarykey;column:item_id" json:"item_id"`
	OrderID  uint64    `gorm:"column:order_id" json:"order_id"`
	Quantity uint      `json:"quantity"`
	Price    uint      `json:"price"`
}
