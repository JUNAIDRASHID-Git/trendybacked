package adminController

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"gorm.io/gorm"
)

// ListPendingAdmins returns all admins awaiting approval.
func ListPendingAdmins(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var pending []models.Admin
		if err := db.Where("approved = ?", false).Find(&pending).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending admins"})
			return
		}
		c.JSON(http.StatusOK, pending)
	}
}

func ApproveAdmin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		var admin models.Admin
		if err := db.Where("email = ?", req.Email).First(&admin).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Admin not found"})
			return
		}

		if err := db.Model(&admin).Update("approved", true).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve admin"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Admin approved"})
	}
}

func RejectAdmin(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			Email string `json:"email"`
		}
		if err := c.ShouldBindJSON(&req); err != nil || req.Email == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		if err := db.Where("email = ?", req.Email).Delete(&models.Admin{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject admin"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Admin rejected"})
	}
}
