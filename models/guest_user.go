package models

import "time"

type GuestUser struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	ExpiresAt time.Time `json:"expires_at"`
}
