package models

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID           uint           `gorm:"primaryKey;autoIncrement"`
	Name         string         `gorm:"not null"`
	Description  string
	SalePrice    float64        `gorm:"not null"`
	RegularPrice float64
	BaseCost     float64
	Image        string         `gorm:"not null"`
	Weight       float64        `gorm:"not null"`
	Categories   []Category     `gorm:"many2many:product_categories;"` // 💡 Many-to-many relationship
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}
