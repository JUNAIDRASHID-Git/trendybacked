package models

import "time"

type User struct {
	ID        string `gorm:"primaryKey" json:"id"`
	Email     string `gorm:"unique;not null"`
	Phone     string
	Name      string
	Picture   string
	Provider  string
	Address   Address `gorm:"embedded"`          // Embeds address fields directly
	CartItems []Cart  `gorm:"foreignKey:UserID"` // One-to-many relationship
	Orders    []Order `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL" json:"orders"`
	CreatedAt time.Time
}

// Address model embedded in User
type Address struct {
	Country    string
	State      string
	City       string
	Street     string
	PostalCode string
}
