package models

type Category struct {
	ID       uint      `gorm:"primaryKey;autoIncrement"`
	Name     string    `gorm:"unique;not null"`
	Products []Product `gorm:"many2many:product_categories"`
}
