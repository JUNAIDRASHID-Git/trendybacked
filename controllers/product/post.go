package productcontroller

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"gorm.io/gorm"
)

// CreateProduct creates a new product with multiple categories + image upload.
func CreateProduct(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Required fields
		ename := c.PostForm("ename")
		salePriceStr := c.PostForm("sale_price")
		weightStr := c.PostForm("weight")
		if ename == "" || salePriceStr == "" || weightStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ename, sale_price, and weight are required"})
			return
		}

		// Optional fields
		arname := c.PostForm("arname")
		edescription := c.PostForm("edescription")
		ardescription := c.PostForm("ardescription")
		regularPriceStr := c.PostForm("regular_price")
		baseCostStr := c.PostForm("base_cost")
		categoryIDsStr := c.PostForm("category_ids")

		// Convert numerics
		salePrice, err := strconv.ParseFloat(salePriceStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sale_price"})
			return
		}
		weight, err := strconv.ParseFloat(weightStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid weight"})
			return
		}

		var regularPrice, baseCost float64
		if regularPriceStr != "" {
			if rp, parseErr := strconv.ParseFloat(regularPriceStr, 64); parseErr == nil {
				regularPrice = rp
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid regular_price"})
				return
			}
		}
		if baseCostStr != "" {
			if bc, parseErr := strconv.ParseFloat(baseCostStr, 64); parseErr == nil {
				baseCost = bc
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid base_cost"})
				return
			}
		}

		// Categories
		var categories []models.Category
		if categoryIDsStr != "" {
			idTokens := strings.Split(categoryIDsStr, ",")
			var parsedIDs []uint
			for _, tok := range idTokens {
				tok = strings.TrimSpace(tok)
				if tok == "" {
					continue
				}
				if id64, parseErr := strconv.ParseUint(tok, 10, 64); parseErr == nil {
					parsedIDs = append(parsedIDs, uint(id64))
				} else {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category_ids format"})
					return
				}
			}
			if len(parsedIDs) > 0 {
				if err := db.Where("id IN ?", parsedIDs).Find(&categories).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
					return
				}
			}
		}

		// Image upload
		file, err := c.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Image is required"})
			return
		}
		filename := strings.ReplaceAll(file.Filename, " ", "_")

		// 🔥 Use absolute path
		saveDir := "/var/www/trendybacked/uploads/products"
		if err := os.MkdirAll(saveDir, os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create upload folder: %v", err)})
			return
		}
		savePath := filepath.Join(saveDir, filename)

		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save image: %v", err)})
			return
		}

		// Public URL (served by nginx/gin)
		imageURL := fmt.Sprintf("/uploads/products/%s", filename)

		// Transaction
		tx := db.Begin()
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}

		newProduct := models.Product{
			EName:         ename,
			ARName:        arname,
			EDescription:  edescription,
			ARDescription: ardescription,
			SalePrice:     salePrice,
			RegularPrice:  regularPrice,
			BaseCost:      baseCost,
			Weight:        weight,
			Image:         imageURL,
			Categories:    categories,
		}

		if err := tx.Create(&newProduct).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
			return
		}
		if err := tx.Commit().Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		c.JSON(http.StatusCreated, newProduct)
	}
}
