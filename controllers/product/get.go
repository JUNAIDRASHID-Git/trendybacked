package productcontroller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"gorm.io/gorm"
)

// GetProductByID returns a single product (with its categories).
// URL param: /products/:id
func GetProductByID(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		if idParam == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
			return
		}

		// Convert string to int
		id, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}

		var product models.Product
		// Attempt to retrieve the product, preload categories
		if err := db.Preload("Categories").First(&product, id).Error; err != nil {
			// Check for specific errors
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve product"})
			}
			return
		}

		// Return product details in the response
		c.JSON(http.StatusOK, product)
	}
}
