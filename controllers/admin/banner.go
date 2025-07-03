package adminController

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"github.com/junaidrashid-git/ecommerce-api/utils"
	"gorm.io/gorm"
)

func UploadBanner(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, fileHeader, err := c.Request.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No image uploaded"})
			return
		}
		defer file.Close()

		imageURL, err := utils.UploadImage(file, fileHeader)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cloudinary upload failed", "details": err.Error()})
			return
		}

		banner := models.Banner{ImageURL: imageURL}
		if err := db.Create(&banner).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB save failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Banner uploaded", "data": banner})
	}
}

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

func DeleteBanner(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get banner ID from path
		id := c.Param("id")
		var banner models.Banner

		// Find banner
		if err := db.First(&banner, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Banner not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Extract public ID from Cloudinary URL (assuming URL format)
		publicID := extractCloudinaryPublicID(banner.ImageURL)
		if publicID == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid Cloudinary URL"})
			return
		}

		// Delete image from Cloudinary
		if err := utils.DeleteImage(publicID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Cloudinary delete failed", "details": err.Error()})
			return
		}

		// Delete record from database
		if err := db.Delete(&banner).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete from database"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Banner deleted"})
	}
}

func extractCloudinaryPublicID(imageURL string) string {
	// Find the "/upload/" part, Cloudinary always includes it
	uploadIndex := strings.Index(imageURL, "/upload/")
	if uploadIndex == -1 {
		return ""
	}

	// Get the path after "/upload/"
	path := imageURL[uploadIndex+len("/upload/"):]

	// Remove version info like v1234567890/
	pathParts := strings.SplitN(path, "/", 2)
	if len(pathParts) < 2 {
		return ""
	}

	publicPath := pathParts[1]
	publicID := strings.TrimSuffix(publicPath, ".jpg")
	publicID = strings.TrimSuffix(publicID, ".png")
	publicID = strings.TrimSuffix(publicID, ".webp")
	publicID = strings.TrimSuffix(publicID, ".jpeg")

	return publicID
}
