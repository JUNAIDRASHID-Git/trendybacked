package qrcontroller

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/junaidrashid-git/ecommerce-api/models"
)

// HandleQRFileUpload handles file uploads and saves info to DB
func HandleQRFileUpload(db *gorm.DB, uploadDir string, publicBaseURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse uploaded file
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
			return
		}

		// Sanitize filename
		re := regexp.MustCompile(`[^\w\d\-_\.]`)
		cleanName := re.ReplaceAllString(file.Filename, "_")
		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), cleanName)

		// Ensure upload directory exists
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to create upload folder: %v", err),
			})
			return
		}

		// Save file to disk
		savePath := filepath.Join(uploadDir, filename)
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to save file: %v", err),
			})
			return
		}

		// Construct public URL
		fileURL := fmt.Sprintf("%s/qrfiles/%s", publicBaseURL, filename)

		// Save record in database
		qrFile, err := models.SaveQRFile(db, filename, fileURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to save file record: %v", err),
			})
			return
		}

		// Log and respond
		log.Printf("✅ QR file uploaded & saved: %s -> %s", filename, fileURL)

		c.JSON(http.StatusOK, gin.H{
			"message":  "File uploaded and saved successfully",
			"id":       qrFile.ID,
			"file_url": qrFile.FileURL,
		})
	}
}

func GetAllQRFilesHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		files, err := models.GetAllQRFiles(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch QR files"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": files})
	}
}
