package model

import (
	"time"
)

type Order struct {
	OrderID     int64      `gorm:"primarykey;column:order_id" json:"order_id"`
	CustomerID  int64      `json:"customer_id"`
	LineItems   []LineItem `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE" json:"line_items"`
	CreatedAt   *time.Time `json:"created_at"`
	ShippedAt   *time.Time `json:"shipped_at"`
	CompletedAt *time.Time `json:"completed_at"`
}
