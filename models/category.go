package models

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID        uint           `gorm:"primaryKey;autoIncrement"`
	Name      string         `gorm:"unique;not null"`
	Products  []Product      `gorm:"many2many:product_categories;" json:"products"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// Constructor/helper to initialize an empty Products array
func NewCategory(name string) *Category {
	return &Category{
		Name:     name,
		Products: []Product{},
	}
}
