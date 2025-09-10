package models

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID            uint    `gorm:"primaryKey;autoIncrement"`
	EName         string  `gorm:"not null"` // English Name
	ARName        string  // Arabic Name
	EDescription  string  // English Description
	ARDescription string  // Arabic Description
	SalePrice     float64 `gorm:"not null"` // Required
	RegularPrice  float64
	BaseCost      float64
	Image         string     `gorm:"not null"`
	Weight        float64    `gorm:"not null"` // Required
	Categories    []Category `gorm:"many2many:product_categories;"`
	Stock         int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}
