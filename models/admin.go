package models

type Admin struct {
	ID       uint   `gorm:"primaryKey"`
	Email    string `gorm:"unique"`
	Name     string
	Picture  string
	Approved bool
}
