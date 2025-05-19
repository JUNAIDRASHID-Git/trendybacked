package models

import (
	"time"
	"gorm.io/gorm"
)

type Category struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"unique;not null"`
	Products  []Product `gorm:"many2many:product_categories;"` // Optional: reverse relation
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
