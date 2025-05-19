package controllers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"gorm.io/gorm"
)

// CreateProduct handles the creation of a new product with multiple categories.
func CreateProduct(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.PostForm("name")
		description := c.PostForm("description")
		salePriceStr := c.PostForm("sale_price")
		regularPriceStr := c.PostForm("regular_price")
		baseCostStr := c.PostForm("base_cost")
		weightStr := c.PostForm("weight")
		categoryIDsStr := c.PostForm("category_ids") // Comma-separated category IDs (e.g., "1,3,5")

		if name == "" || salePriceStr == "" || weightStr == "" || categoryIDsStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Name, Sale Price, Weight, and Categories are required"})
			return
		}

		// Parse numeric fields
		salePrice, err := strconv.ParseFloat(salePriceStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sale price"})
			return
		}

		weight, err := strconv.ParseFloat(weightStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid weight"})
			return
		}

		var regularPrice float64
		if regularPriceStr != "" {
			regularPrice, err = strconv.ParseFloat(regularPriceStr, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid regular price"})
				return
			}
		}

		var baseCost float64
		if baseCostStr != "" {
			baseCost, err = strconv.ParseFloat(baseCostStr, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid base cost"})
				return
			}
		}

		// Parse category IDs
		var categoryIDs []uint
		for _, idStr := range strings.Split(categoryIDsStr, ",") {
			id, err := strconv.ParseUint(strings.TrimSpace(idStr), 10, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID: " + idStr})
				return
			}
			categoryIDs = append(categoryIDs, uint(id))
		}

		// Get categories
		var categories []models.Category
		if err := db.Where("id IN ?", categoryIDs).Find(&categories).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
			return
		}

		// Handle image
		file, err := c.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Image is required"})
			return
		}
		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), filepath.Base(file.Filename))
		filePath := "uploads/" + filename
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
			return
		}

		product := models.Product{
			Name:         name,
			Description:  description,
			SalePrice:    salePrice,
			RegularPrice: regularPrice,
			BaseCost:     baseCost,
			Weight:       weight,
			Image:        "/" + filePath,
			Categories:   categories,
		}

		if err := db.Create(&product).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
			return
		}

		c.JSON(http.StatusCreated, product)
	}
}

// DeleteProduct handles deleting a product by ID.
func DeleteProduct(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var product models.Product

		if err := db.First(&product, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		if err := db.Delete(&product).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
	}
}

// GetProducts fetches all products including their categories.
func GetProducts(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var products []models.Product

		if err := db.Preload("Categories").Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
			return
		}

		c.JSON(http.StatusOK, products)
	}
}
