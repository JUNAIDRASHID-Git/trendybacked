package productcontroller

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"gorm.io/gorm"
)

// CreateCategory creates a new category.
func CreateCategory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ename := c.PostForm("ename")
		arname := c.PostForm("arname")

		if ename == "" || arname == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ename and arname are required"})
			return
		}

		// Optional image upload
		var imageURL string
		file, err := c.FormFile("image")
		if err == nil { // Image is optional
			uploadDir := "/var/www/trendybacked/uploads/categories"
			os.MkdirAll(uploadDir, os.ModePerm)

			filename := strings.ReplaceAll(file.Filename, " ", "_")
			savePath := filepath.Join(uploadDir, filename)

			if err := c.SaveUploadedFile(file, savePath); err != nil {
				c.JSON(500, gin.H{"error": "Failed to save image"})
				return
			}

			imageURL = "/uploads/categories/" + filename
		}

		category := models.Category{
			EName:  ename,
			ARName: arname,
			Image:  imageURL,
		}

		if err := db.Create(&category).Error; err != nil {
			c.JSON(500, gin.H{"error": "Failed to create category"})
			return
		}

		c.JSON(201, category)
	}
}

func GetAllCategoriesWithProducts(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var categories []models.Category

		// Preload Products
		if err := db.Preload("Products").Find(&categories).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories with products"})
			return
		}

		c.JSON(http.StatusOK, categories)
	}
}

func GetCategoryByID(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var category models.Category
		if err := db.Preload("Products").First(&category, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
		}

		c.JSON(http.StatusOK, category)
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

		var category models.Category
		if err := db.First(&category, id).Error; err != nil {
			c.JSON(404, gin.H{"error": "Category not found"})
			return
		}

		// Update names if provided
		if v := c.PostForm("ename"); v != "" {
			category.EName = v
		}
		if v := c.PostForm("arname"); v != "" {
			category.ARName = v
		}

		// Optional image upload
		file, err := c.FormFile("image")
		if err == nil {
			uploadDir := "/var/www/trendybacked/uploads/categories"
			os.MkdirAll(uploadDir, os.ModePerm)

			filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(),
				strings.ReplaceAll(file.Filename, " ", "_"))

			savePath := filepath.Join(uploadDir, filename)

			if err := c.SaveUploadedFile(file, savePath); err != nil {
				c.JSON(500, gin.H{"error": "Failed to save image"})
				return
			}

			category.Image = "/uploads/categories/" + filename
		}

		if err := db.Save(&category).Error; err != nil {
			c.JSON(500, gin.H{"error": "Failed to update category"})
			return
		}

		c.JSON(200, category)
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
