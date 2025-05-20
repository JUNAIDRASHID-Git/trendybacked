package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

func ImportProductsFromExcel(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Excel file is required"})
			return
		}

		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
			return
		}
		defer src.Close()

		f, err := excelize.OpenReader(src)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Excel file"})
			return
		}

		rows, err := f.GetRows("Sheet1")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read sheet 'Sheet1'"})
			return
		}

		if len(rows) <= 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No data found in Excel"})
			return
		}

		var products []models.Product

		for i, row := range rows {
			if i == 0 {
				continue // Skip header row
			}

			// Require at least: name (0), sale_price (2), weight (4)
			if len(row) < 5 || strings.TrimSpace(row[0]) == "" || strings.TrimSpace(row[2]) == "" || strings.TrimSpace(fmt.Sprintf("%v", row[4])) == "" {
				continue
			}

			name := strings.TrimSpace(row[0])
			description := ""
			if len(row) > 1 {
				description = strings.TrimSpace(row[1])
			}

			salePrice, err := strconv.ParseFloat(strings.TrimSpace(row[2]), 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid sale price on row %d", i+1)})
				return
			}

			regularPrice := 0.0
			if len(row) > 3 && row[3] != "" {
				regularPrice, err = strconv.ParseFloat(strings.TrimSpace(row[3]), 64)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid regular price on row %d", i+1)})
					return
				}
			}

			weight, err := strconv.ParseFloat(fmt.Sprintf("%v", row[4]), 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid weight on row %d", i+1)})
				return
			}

			image := "/images/default_product_image.jpg"
			if len(row) > 5 && strings.TrimSpace(row[5]) != "" {
				image = strings.TrimSpace(row[5])
			}

			baseCost := 0.0 // You can extend this to read from Excel later

			var categories []models.Category
			if len(row) > 6 && strings.TrimSpace(row[6]) != "" {
				categoryIDs := strings.Split(row[6], ",")
				for _, idStr := range categoryIDs {
					idStr = strings.TrimSpace(idStr)
					if idStr == "" {
						continue
					}
					catID, err := strconv.ParseUint(idStr, 10, 64)
					if err != nil {
						continue
					}
					var category models.Category
					if err := db.First(&category, catID).Error; err == nil {
						categories = append(categories, category)
					}
				}
			}

			product := models.Product{
				Name:         name,
				Description:  description,
				SalePrice:    salePrice,
				RegularPrice: regularPrice,
				BaseCost:     baseCost,
				Image:        image,
				Weight:       weight,
				Categories:   categories,
			}

			products = append(products, product)
		}

		if len(products) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No valid products found to import"})
			return
		}

		for _, product := range products {
			if err := db.Create(&product).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save product: %s", product.Name)})
				return
			}
			if len(product.Categories) > 0 {
				if err := db.Model(&product).Association("Categories").Replace(product.Categories); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to associate categories"})
					return
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Products imported successfully",
			"count":   len(products),
		})
	}
}
