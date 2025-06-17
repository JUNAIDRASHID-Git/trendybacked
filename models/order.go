package models

import "time"

type OrderStatus string

const (
	StatusPending  OrderStatus = "pending"
	StatusSuccess  OrderStatus = "success"
	StatusCanceled OrderStatus = "canceled"
)

type Order struct {
	ID            uint        `gorm:"primaryKey" json:"id"`
	UserID        string      `gorm:"not null" json:"user_id"`
	User          User        `gorm:"foreignKey:UserID" json:"user"`
	Items         []OrderItem `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE" json:"items"`
	ShippingCost  float64     `json:"shipping_cost"`
	TotalAmount   float64     `json:"total_amount"`
	Status        OrderStatus `gorm:"default:'pending'" json:"status"`
	PaymentMethod string      `json:"payment_method"` // Optional
	CreatedAt     time.Time   `json:"created_at"`
}

type OrderItem struct {
	ID                  uint `gorm:"primaryKey"`
	OrderID             uint `gorm:"index"`
	ProductID           uint
	ProductEName        string
	ProductArName       string
	ProductImage        string
	ProductSalePrice    float64
	ProductRegularPrice float64
	Weight              float64
	Quantity            int
}
