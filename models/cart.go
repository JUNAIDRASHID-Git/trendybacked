package models

import "time"

type Cart struct {
	CartID    uint       `gorm:"primaryKey"`
	UserID    string     `gorm:"uniqueIndex"`                                   // Enforces ONE cart per user
	Items     []CartItem `gorm:"foreignKey:CartID;constraint:OnDelete:CASCADE"` // Cascade delete items if cart is deleted
	CreatedAt time.Time
	UpdatedAt time.Time
}

type CartItem struct {
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
