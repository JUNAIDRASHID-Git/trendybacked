package adminController

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"gorm.io/gorm"
)

func GetAllAdmins(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var admins []models.Admin

		if err := db.Find(&admins).Error; err != nil {
			log.Println("❌ Failed to fetch admins:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch admins"})
			return
		}

		c.JSON(http.StatusOK, admins)
	}
}
