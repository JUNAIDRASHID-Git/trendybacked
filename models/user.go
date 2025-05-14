package models

import "time"

type User struct {
	ID        string `gorm:"primaryKey"`
	Email     string `gorm:"unique;not null"`
	Name      string
	Picture   string
	Provider  string
	CreatedAt time.Time
}
