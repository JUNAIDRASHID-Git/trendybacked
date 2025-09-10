package productcontroller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"gorm.io/gorm"
)

func GetProducts(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1️⃣ Filtering & sorting params
		search := c.Query("search")
		categoryID := c.Query("category_id")
		minPriceStr := c.Query("min_price")
		maxPriceStr := c.Query("max_price")
		sortBy := c.DefaultQuery("sort_by", "created_at")
		sortOrder := strings.ToLower(c.DefaultQuery("order", "desc"))
		if sortOrder != "asc" && sortOrder != "desc" {
			sortOrder = "desc"
		}

		// 2️⃣ Build base query
		query := db.Model(&models.Product{}).Preload("Categories")

		// 3️⃣ Apply search filter
		if search != "" {
			likePattern := "%" + search + "%"
			query = query.Where(`
				e_name ILIKE ? OR e_description ILIKE ? OR ar_name ILIKE ? OR ar_description ILIKE ?
			`, likePattern, likePattern, likePattern, likePattern)
		}

		// 4️⃣ Apply price range filter
		if minPriceStr != "" {
			if mp, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
				query = query.Where("sale_price >= ?", mp)
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid min_price"})
				return
			}
		}
		if maxPriceStr != "" {
			if mp, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
				query = query.Where("sale_price <= ?", mp)
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid max_price"})
				return
			}
		}

		// 5️⃣ Apply category filter
		if categoryID != "" {
			if cid, err := strconv.ParseUint(categoryID, 10, 64); err == nil {
				query = query.
					Joins("JOIN product_categories pc ON pc.product_id = products.id").
					Where("pc.category_id = ?", uint(cid))
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category_id"})
				return
			}
		}

		// 6️⃣ Apply sorting
		orderClause := fmt.Sprintf("%s %s", sortBy, sortOrder)
		var products []models.Product
		if err := query.Order(orderClause).Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
			return
		}
		// 8️⃣ Return products
		c.JSON(http.StatusOK, products)
	}
}
