package adminController

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"gorm.io/gorm"
)

const uploadDir = "/var/www/trendybacked/uploads/products" // Full path to server folder
const domain = "https://api.trendy-c.com"                  // Your domain

// EnsureBannerTable - auto-migrate Banner table
func EnsureBannerTable(db *gorm.DB) {
	if err := db.AutoMigrate(&models.Banner{}); err != nil {
		fmt.Println("❌ Failed to migrate Banner table:", err)
	} else {
		fmt.Println("✅ Banner table ready")
	}
}

// UploadBanner - Save image locally and store full URL in DB
func UploadBanner(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, fileHeader, err := c.Request.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No image uploaded"})
			return
		}
		defer file.Close()

		// Ensure upload directory exists
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload folder"})
			return
		}

		// Original filename
		origName := fileHeader.Filename
		ext := filepath.Ext(origName)                 // ".jpg"
		baseName := strings.TrimSuffix(origName, ext) // remove last extension

		// Remove duplicate extensions like ".jpg.jpg"
		for {
			e := filepath.Ext(baseName)
			if e != "" && (e == ".jpg" || e == ".jpeg" || e == ".png" || e == ".gif") {
				baseName = strings.TrimSuffix(baseName, e)
			} else {
				break
			}
		}

		// Clean spaces
		baseName = strings.ReplaceAll(baseName, " ", "_")

		// Final filename
		newFileName := fmt.Sprintf("%d_%s%s", time.Now().Unix(), baseName, ext)
		savePath := filepath.Join(uploadDir, newFileName)

		// Save file locally
		if err := c.SaveUploadedFile(fileHeader, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}

		// Full URL including domain
		imageURL := fmt.Sprintf("%s/uploads/products/%s", domain, newFileName)

		// Save banner in DB
		banner := models.Banner{ImageURL: imageURL}
		if err := db.Create(&banner).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB save failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Banner uploaded", "data": banner})
	}
}

// GetBanners - List banners
func GetBanners(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var banners []models.Banner
		if err := db.Find(&banners).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get banners"})
			return
		}
		c.JSON(http.StatusOK, banners)
	}
}

// DeleteBanner - Delete both DB record & local file
func DeleteBanner(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var banner models.Banner

		// Find banner in DB
		if err := db.First(&banner, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Banner not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Delete file from disk
		if banner.ImageURL != "" {
			// Convert URL to local path
			localPath := strings.Replace(banner.ImageURL, domain, "/var/www/trendybacked", 1)
			if err := os.Remove(localPath); err != nil && !os.IsNotExist(err) {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file"})
				return
			}
		}

		// Delete DB record
		if err := db.Delete(&banner).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete from database"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Banner deleted"})
	}
}
