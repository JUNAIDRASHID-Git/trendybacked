package models

import "time"

// GuestCart represents a cart for guest users
type GuestCart struct {
	CartID    uint            `gorm:"primaryKey"`
	GuestID   string          `gorm:"uniqueIndex"`                                   // Enforces ONE cart per guest
	Items     []GuestCartItem `gorm:"foreignKey:CartID;constraint:OnDelete:CASCADE"` // Cascade delete items if cart is deleted
	CreatedAt time.Time
	UpdatedAt time.Time
}

// GuestCartItem represents items in the guest cart
type GuestCartItem struct {
	ID                  uint `gorm:"primaryKey"`
	CartID              uint `gorm:"index"` // Faster queries
	ProductID           uint
	ProductEName        string // English name of the product
	ProductArName       string // Arabic name of the product
	ProductImage        string
	ProductStock        int
	ProductSalePrice    float64
	ProductRegularPrice float64
	Weight              float64
	Quantity            int
	AddedAt             time.Time
}
