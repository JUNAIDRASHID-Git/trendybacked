package productcontroller

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"gorm.io/gorm"
)

const uploadDir = "/var/www/trendybacked/uploads/products"

// UpdateProduct updates an existing product by ID.
// Accepts the same fields as CreateProduct and an optional "image" file.
func UpdateProduct(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get product ID from URL
		idStr := c.Param("id")
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
			return
		}

		// Fetch existing product
		var product models.Product
		if err := db.Preload("Categories").First(&product, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		// Helper to parse float fields safely
		parseFloat := func(val string) *float64 {
			if val == "" {
				return nil
			}
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				return &f
			}
			return nil
		}

		// Helper to parse int fields safely
		parseInt := func(val string) *int {
			if val == "" {
				return nil
			}
			if i, err := strconv.Atoi(val); err == nil {
				return &i
			}
			return nil
		}

		// Parse form fields (optional updates)
		if v := c.PostForm("ename"); v != "" {
			product.EName = v
		}
		if v := c.PostForm("arname"); v != "" {
			product.ARName = v
		}
		if v := c.PostForm("edescription"); v != "" {
			product.EDescription = v
		}
		if v := c.PostForm("ardescription"); v != "" {
			product.ARDescription = v
		}
		if v := parseFloat(c.PostForm("sale_price")); v != nil {
			product.SalePrice = *v
		}
		if v := parseFloat(c.PostForm("regular_price")); v != nil {
			product.RegularPrice = *v
		}
		if v := parseFloat(c.PostForm("base_cost")); v != nil {
			product.BaseCost = *v
		}
		if v := parseFloat(c.PostForm("weight")); v != nil {
			product.Weight = *v
		}
		if v := parseInt(c.PostForm("stock")); v != nil { // 👈 Stock support
			product.Stock = *v
		}

		// Update categories if provided
		if categoryIDsStr := c.PostForm("category_ids"); categoryIDsStr != "" {
			idTokens := strings.Split(categoryIDsStr, ",")
			var parsedIDs []uint
			for _, tok := range idTokens {
				tok = strings.TrimSpace(tok)
				if tok == "" {
					continue
				}
				if id64, parseErr := strconv.ParseUint(tok, 10, 64); parseErr == nil {
					parsedIDs = append(parsedIDs, uint(id64))
				}
			}
			if len(parsedIDs) > 0 {
				var categories []models.Category
				if err := db.Where("id IN ?", parsedIDs).Find(&categories).Error; err == nil {
					product.Categories = categories
				}
			}
		}

		// Handle optional image upload
		file, err := c.FormFile("image")
		if err == nil {
			// Ensure upload directory exists
			if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create upload folder"})
				return
			}

			// Create unique filename
			ext := filepath.Ext(file.Filename)
			base := strings.TrimSuffix(filepath.Base(file.Filename), ext)
			base = strings.ReplaceAll(base, " ", "_")
			filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), base, ext)

			savePath := filepath.Join(uploadDir, filename)

			if err := c.SaveUploadedFile(file, savePath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
				return
			}

			// Save relative path for client access
			product.Image = fmt.Sprintf("/uploads/products/%s", filename)
		}

		// Save updated product
		if err := db.Save(&product).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
			return
		}

		c.JSON(http.StatusOK, product)
	}
}
