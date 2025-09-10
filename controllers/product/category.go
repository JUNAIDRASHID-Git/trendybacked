package productcontroller

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"gorm.io/gorm"
)

// CreateCategory creates a new category.
func CreateCategory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input struct {
			EName  string `json:"ename" binding:"required"`
			ARName string `json:"arname" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "English and Arabic names are required"})
			return
		}

		newCategory := models.Category{
			EName:    input.EName,
			ARName:   input.ARName,
			Products: []models.Product{},
		}

		if err := db.Create(&newCategory).Error; err != nil {
			if strings.Contains(err.Error(), "unique") {
				c.JSON(http.StatusConflict, gin.H{"error": "Category already exists"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category"})
			}
			return
		}

		c.JSON(http.StatusCreated, newCategory)
	}
}

// GetAllCategories returns all categories.
func GetAllCategories(db *gorm.DB) gin.HandlerFunc {
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

// UpdateCategory updates an existing category.
func UpdateCategory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Category ID is required"})
			return
		}

		var input struct {
			EName  string `json:"ename" binding:"required"`
			ARName string `json:"arname" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Both English and Arabic names are required"})
			return
		}

		result := db.Model(&models.Category{}).
			Where("id = ?", id).
			Updates(models.Category{EName: input.EName, ARName: input.ARName})

		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
			return
		}
		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
		}

		var updated models.Category
		if err := db.First(&updated, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated category"})
			return
		}

		c.JSON(http.StatusOK, updated)
	}
}

// DeleteCategory deletes a category and clears product associations.
func DeleteCategory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Category ID is required"})
			return
		}

		var cat models.Category
		if err := db.Preload("Products").First(&cat, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
		}

		tx := db.Begin()
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
			}
		}()

		if err := tx.Model(&cat).Association("Products").Clear(); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear product associations"})
			return
		}

		if err := tx.Delete(&cat).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
			return
		}

		if err := tx.Commit().Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
	}
}
