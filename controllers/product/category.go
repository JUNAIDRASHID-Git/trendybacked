package productcontroller

import (
	"net/http"
	"strings"
	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"gorm.io/gorm"
)

// CreateCategory creates a new category. JSON body: { "name": "Category Name" }.
// Returns 201 + created category, or appropriate error.
func CreateCategory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1️⃣ Bind input
		var input struct {
			Name string `json:"name" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
			return
		}

		// 2️⃣ Initialize category (no Products yet)
		newCategory := models.Category{
			Name:     input.Name,
			Products: []models.Product{},
		}

		// 3️⃣ Save to DB
		if err := db.Create(&newCategory).Error; err != nil {
			// Check for unique constraint violation (duplicate name)
			if strings.Contains(err.Error(), "unique") {
				c.JSON(http.StatusConflict, gin.H{"error": "Category already exists"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create category"})
			return
		}

		// 4️⃣ Return created category
		c.JSON(http.StatusCreated, newCategory)
	}
}

// GetAllCategories returns all categories (no pagination).
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

// UpdateCategory updates only the Name of an existing category. JSON body: { "name": "New Name" }.
// Returns 200 + updated category, or 404/400/500 appropriately.
func UpdateCategory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1️⃣ Parse category ID from URL
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Category ID is required"})
			return
		}

		// 2️⃣ Bind input
		var input struct {
			Name string `json:"name" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
			return
		}

		// 3️⃣ Try updating
		result := db.Model(&models.Category{}).
			Where("id = ?", id).
			Updates(models.Category{Name: input.Name})

		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category"})
			return
		}
		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
		}

		// 4️⃣ Fetch updated category
		var updated models.Category
		if err := db.First(&updated, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated category"})
			return
		}

		c.JSON(http.StatusOK, updated)
	}
}

// DeleteCategory removes a category by ID.
// It also clears any product-category associations (join table entries), then deletes.
func DeleteCategory(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1️⃣ Parse category ID
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Category ID is required"})
			return
		}

		// 2️⃣ Fetch the category (to clear associations)
		var cat models.Category
		if err := db.Preload("Products").First(&cat, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
			return
		}

		// 3️⃣ Start a transaction: clear associations, then delete
		tx := db.Begin()
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}

		// 4️⃣ Clear association in join table
		if err := tx.Model(&cat).Association("Products").Clear(); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear category-product associations"})
			return
		}

		// 5️⃣ Delete the category itself
		if err := tx.Delete(&cat).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
			return
		}

		// 6️⃣ Commit
		if err := tx.Commit().Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit deletion"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Category deleted successfully"})
	}
}
