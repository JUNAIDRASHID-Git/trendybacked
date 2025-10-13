package models

import "time"

type OrderStatus string
type PaymentStatus string

const (
	// Order statuses (typical e-commerce flow)
	OrderStatusPending     OrderStatus = "pending"       // Order placed, awaiting confirmation
	OrderStatusConfirmed   OrderStatus = "confirmed"     // Confirmed by seller
	OrderStatusReadyToShip OrderStatus = "ready_to_ship" // Packed and ready for dispatch
	OrderStatusShipped     OrderStatus = "shipped"       // Out for delivery
	OrderStatusDelivered   OrderStatus = "delivered"     // Customer received the item
	OrderStatusReturned    OrderStatus = "returned"      // Customer returned the item
	OrderStatusCancelled   OrderStatus = "cancelled"     // Cancelled before shipping

	// Payment statuses
	PaymentStatusPending  PaymentStatus = "pending"  // Payment not completed yet
	PaymentStatusPaid     PaymentStatus = "paid"     // Payment completed successfully
	PaymentStatusFailed   PaymentStatus = "failed"   // Payment attempt failed
	PaymentStatusRefunded PaymentStatus = "refunded" // Money returned to customer
)

type Order struct {
	ID            uint          `gorm:"primaryKey" json:"id"`
	UserID        string        `gorm:"not null" json:"user_id"`
	User          User          `gorm:"foreignKey:UserID" json:"user"`
	Items         []OrderItem   `gorm:"foreignKey:OrderID;constraint:OnDelete:CASCADE" json:"items"`
	ShippingCost  float64       `json:"shipping_cost"`
	TotalAmount   float64       `json:"total_amount"`
	Status        OrderStatus   `gorm:"type:VARCHAR(20);default:'pending'" json:"status"`
	PaymentStatus PaymentStatus `gorm:"type:VARCHAR(20);default:'pending'" json:"payment_status"`
	PaymentMethod string        `json:"payment_method"` // e.g. "card", "cod"
	CreatedAt     time.Time     `json:"created_at"`
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
