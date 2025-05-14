package models

import (
	"gorm.io/gorm"
	"time"
)

type Product struct {
	ID           uint    `gorm:"primaryKey;autoIncrement"`
	Name         string  `gorm:"not null"`
	Description  string  `gorm:""`
	SalePrice    float64 `gorm:"not null"`
	RegularPrice float64 `gorm:"not null"`
	BaseCost     float64 `gorm:"not null"`
	Image        string  `gorm:"not null"`
	Weight       float64 `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}
