package controllers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"gorm.io/gorm"
)

// CreateProduct handles the creation of a new product.
func CreateProduct(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.PostForm("name")
		description := c.PostForm("description")
		salePriceStr := c.PostForm("sale_price")
		regularPriceStr := c.PostForm("regular_price")
		baseCostStr := c.PostForm("base_cost")
		weightStr := c.PostForm("weight")

		// Validate required fields
		if name == "" || salePriceStr == "" || weightStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Name, Sale Price, and Weight are required"})
			return
		}

		// Parse and validate required fields
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

		// Parse optional fields
		var regularPrice, baseCost float64

		if regularPriceStr != "" {
			regularPrice, err = strconv.ParseFloat(regularPriceStr, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid regular price"})
				return
			}
		}

		if baseCostStr != "" {
			baseCost, err = strconv.ParseFloat(baseCostStr, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid base cost"})
				return
			}
		}

		// Handle image upload
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
		}

		if err := db.Create(&product).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save product"})
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

// GetProducts fetches all products.
func GetProducts(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var products []models.Product

		if err := db.Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
			return
		}

		c.JSON(http.StatusOK, products)
	}
}
