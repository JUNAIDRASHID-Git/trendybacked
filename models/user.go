package models

import "time"

type User struct {
	ID        string `gorm:"primaryKey"`
	Email     string `gorm:"unique;not null"`
	Phone     string
	Name      string
	Picture   string
	Provider  string
	Address   Address    `gorm:"embedded"`          // Embeds address fields directly
	CartItems []CartItem `gorm:"foreignKey:UserID"` // One-to-many relationship
	CreatedAt time.Time
}

// Address model embedded in User
type Address struct {
	Street     string
	City       string
	State      string
	PostalCode string
	Country    string
}

// CartItem represents an item in the user's cart
type CartItem struct {
	ID        uint `gorm:"primaryKey"`
	UserID    string
	ProductID string // Reference to a Product model
	Quantity  int
	AddedAt   time.Time
}
