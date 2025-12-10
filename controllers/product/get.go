package productcontroller

import (
	"fmt"
	"html"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"gorm.io/gorm"
)

// GetProductByID returns a single product (with its categories and accessible image URL).
// URL param: /products/:id
func GetProductByID(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		if idParam == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
			return
		}

		id, err := strconv.Atoi(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}

		var product models.Product
		if err := db.Preload("Categories").First(&product, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve product"})
			}
			return
		}
		c.JSON(http.StatusOK, product)
	}
}

func GetProductOGHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idParam := c.Param("id")
		if idParam == "" {
			c.String(http.StatusBadRequest, "Product ID is required")
			return
		}

		id, err := strconv.Atoi(idParam)
		if err != nil {
			c.String(http.StatusBadRequest, "Invalid product ID")
			return
		}

		var product models.Product
		if err := db.Preload("Categories").First(&product, id).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.String(http.StatusNotFound, "Product not found")
			} else {
				c.String(http.StatusInternalServerError, "Failed to retrieve product")
			}
			return
		}

		escapedTitle := html.EscapeString(product.EName)
		escapedDescription := html.EscapeString(product.EDescription)
		escapedImage := html.EscapeString("https://server.trendy-c.com" + product.Image)

		htmlContent := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>%s</title>
<meta property="og:title" content="%s" />
<meta property="og:description" content="%s" />
<meta property="og:image" content="%s" />
<meta property="og:url" content="https://trendy-c.com/product/%d" />
<meta property="og:type" content="product" />
<meta property="og:site_name" content="TrendyChef" />
<meta property="og:locale" content="en_US" />
<meta name="twitter:card" content="summary_large_image" />
<meta http-equiv="refresh" content="0;url=https://trendy-c.com/product/%d" />
</head>
<body>
<h1>%s</h1>
<p>%s</p>
<img src="%s" alt="%s" />
</body>
</html>`,
			escapedTitle,       // <title>
			escapedTitle,       // og:title
			escapedDescription, // og:description
			escapedImage,       // og:image
			product.ID,         // og:url
			product.ID,         // meta refresh URL
			escapedTitle,       // <h1>
			escapedDescription, // <p>
			escapedImage,       // <img src>
			escapedTitle,       // <img alt>
		)

		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, htmlContent)
	}
}
