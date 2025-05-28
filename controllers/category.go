package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"gorm.io/gorm"
	"net/http"
)

func CreateCategory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var category models.Category

		// Bind JSON input
		if err := c.ShouldBindJSON(&category); err != nil || category.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
			return
		}

		// Ensure Products is initialized as an empty array
		category.Products = []models.Product{}

		// Save to DB
		if err := db.Create(&category).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category"})
			return
		}

		c.JSON(http.StatusCreated, category)
	}
}

func UpdateCategory(db *gorm.DB) gin.HandlerFunc {

	return func(c *gin.Context) {
		var category models.Category
		if err := c.ShouldBindJSON(&category); err != nil || category.ID == 0 || category.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID and Name is required"})
			return
		}
		if err := db.Model(&models.Category{}).Where("id = ?", category.ID).Updates(category).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to Update Category"})
			return
		}
		c.JSON(http.StatusOK, category)
	}
}

func GetAllCategory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var categories []models.Category
		if err := db.Find(&categories).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
			return
		}
		if len(categories) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"message": "No categories found"})
			return
		}
		c.JSON(http.StatusOK, categories)
	}
}

func DeleteCategory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id") // extract ID from URL param
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
			return
		}

		if err := db.Where("id = ?", id).Delete(&models.Category{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete the category"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
	}
}

