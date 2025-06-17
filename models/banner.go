package models

type Banner struct {
	ID       uint   `gorm:"primaryKey"`
	ImageURL string `gorm:"not null"`
}
