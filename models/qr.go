package models

import (
	"log"
	"time"

	"gorm.io/gorm"
)

type QRFile struct {
	ID        uint           `json:"id" gorm:"primaryKey;autoIncrement"`
	FileName  string         `json:"file_name" gorm:"not null"`
	FileURL   string         `json:"file_url" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func SaveQRFile(db *gorm.DB, fileName, fileURL string) (*QRFile, error) {
	qrFile := &QRFile{
		FileName: fileName,
		FileURL:  fileURL,
	}
	if err := db.Create(qrFile).Error; err != nil {
		return nil, err
	}

	log.Printf("📁 Saved QR file in DB: %s -> %s", fileName, fileURL)
	return qrFile, nil
}

func GetAllQRFiles(db *gorm.DB) ([]QRFile, error) {
	var files []QRFile
	if err := db.Order("created_at DESC").Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}
