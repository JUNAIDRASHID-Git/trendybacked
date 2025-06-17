package adminController

import (
	"net/http"
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
