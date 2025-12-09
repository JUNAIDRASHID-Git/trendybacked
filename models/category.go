package models

type Category struct {
	ID       uint   `gorm:"primaryKey;autoIncrement"`
	EName    string `gorm:"unique;not null"`
	ARName   string `gorm:"unique;not null"`
	Image    string
	Products []Product `gorm:"many2many:product_categories"`
}
