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
)

// HandleQRFileUpload handles file uploads and returns the public URL.
// Production-ready: supports local and prod paths, safe filenames, and logging.
func HandleQRFileUpload(uploadDir string, publicBaseURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse uploaded file
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
			return
		}

		// Sanitize filename: remove any special chars
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

		// Save file
		savePath := filepath.Join(uploadDir, filename)
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to save file: %v", err),
			})
			return
		}

		// Construct public URL
		fileURL := fmt.Sprintf("%s/qrfiles/%s", publicBaseURL, filename)

		// Log upload for monitoring
		log.Printf("QR file uploaded: %s -> %s", file.Filename, fileURL)

		// Respond
		c.JSON(http.StatusOK, gin.H{
			"file_url": fileURL,
			"message":  "File uploaded successfully",
		})
	}
}
