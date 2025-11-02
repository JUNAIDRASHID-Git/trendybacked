package qrcontroller

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"gorm.io/gorm"
)

func DeleteQRFileHandler(db *gorm.DB, uploadDir string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get ID from URL parameter
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
			return
		}

		// Fetch QR file record from DB
		var qrFile models.QRFile
		if err := db.First(&qrFile, "id = ?", id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "QR file not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query QR file"})
			return
		}

		// Delete file from disk
		filePath := filepath.Join(uploadDir, qrFile.FileName)
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file from disk"})
			return
		}

		// Delete record from DB
		if err := db.Delete(&qrFile).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete QR file record"})
			return
		}

		log.Printf("🗑️ QR file deleted: %s", qrFile.FileName)
		c.JSON(http.StatusOK, gin.H{"message": "QR file deleted successfully"})
	}
}
