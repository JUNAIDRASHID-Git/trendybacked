package controllers

import (
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

		// Open uploaded Excel file
		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open file"})
			return
		}
		defer src.Close()

		// Parse Excel
		f, err := excelize.OpenReader(src)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Excel file"})
			return
		}

		rows, err := f.GetRows("Sheet1")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read sheet"})
			return
		}

		var products []models.Product
		for i, row := range rows {
			if i == 0 {
				continue // Skip header
			}
			if len(row) < 6 {
				continue // Skip invalid rows
			}

			name := row[0]
			description := row[1]
			salePrice, _ := strconv.ParseFloat(row[2], 64)
			regularPrice, _ := strconv.ParseFloat(row[3], 64)
			weight, _ := strconv.ParseFloat(row[4], 64)
			image := row[5]
			var categoryIDs []uint

			// Optional: Parse category IDs
			if len(row) > 6 {
				for _, idStr := range strings.Split(row[6], ",") {
					id, err := strconv.ParseUint(strings.TrimSpace(idStr), 10, 64)
					if err == nil {
						categoryIDs = append(categoryIDs, uint(id))
					}
				}
			}

			// Fetch categories
			var categories []models.Category
			if len(categoryIDs) > 0 {
				if err := db.Where("id IN ?", categoryIDs).Find(&categories).Error; err != nil {
					continue // Skip this row on category fetch error
				}
			}

			product := models.Product{
				Name:         name,
				Description:  description,
				SalePrice:    salePrice,
				RegularPrice: regularPrice,
				Weight:       weight,
				Image:        image,
				Categories:   categories,
			}
			products = append(products, product)
		}

		// Bulk insert
		if err := db.Create(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to import products"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Products imported successfully", "count": len(products)})
	}
}
