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
	ID        uint `gorm:"primaryKey"`
	CartID    uint `gorm:"index"` // Faster queries
	ProductID string
	Quantity  int
	AddedAt   time.Time
}
