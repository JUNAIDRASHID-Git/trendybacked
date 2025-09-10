package productcontroller

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"github.com/tealeg/xlsx"
	"gorm.io/gorm"
)

func ImportProductsFromExcel(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		excelFileHeader, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Excel file is required"})
			return
		}

		file, err := excelFileHeader.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open Excel file"})
			return
		}
		defer file.Close()

		xlFile, err := xlsx.OpenReaderAt(file, excelFileHeader.Size)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Excel file"})
			return
		}

		if len(xlFile.Sheets) == 0 || xlFile.Sheets[0].MaxRow < 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Excel file is empty or missing header row"})
			return
		}

		sheet := xlFile.Sheets[0]
		createdCount, updatedCount, skippedCount := 0, 0, 0

		for i := 1; i < sheet.MaxRow; i++ {
			row := sheet.Rows[i]
			if len(row.Cells) < 12 {
				skippedCount++
				continue
			}

			get := func(index int) string {
				if index < len(row.Cells) {
					return strings.TrimSpace(row.Cells[index].String())
				}
				return ""
			}

			idStr := get(0)
			ename := get(1)
			arname := get(2)
			endesc := get(3)
			ardesc := get(4)
			salePrice, err1 := strconv.ParseFloat(get(5), 64)
			regularPrice, _ := strconv.ParseFloat(get(6), 64)
			baseCost, _ := strconv.ParseFloat(get(7), 64)
			weight, err2 := strconv.ParseFloat(get(8), 64)
			stock, _ := strconv.ParseFloat(get(9), 64)
			image := get(10)
			categoryIDStr := get(11)

			if ename == "" || err1 != nil || err2 != nil {
				skippedCount++
				continue
			}

			var categories []models.Category
			for _, part := range strings.Split(categoryIDStr, ",") {
				if id, err := strconv.Atoi(strings.TrimSpace(part)); err == nil {
					categories = append(categories, models.Category{ID: uint(id)})
				}
			}

			product := models.Product{
				EName:         ename,
				ARName:        arname,
				EDescription:  endesc,
				ARDescription: ardesc,
				SalePrice:     salePrice,
				RegularPrice:  regularPrice,
				BaseCost:      baseCost,
				Weight:        weight,
				Stock:         int(stock),
				Image:         image,
				Categories:    categories,
			}

			if idStr != "" {
				if id, err := strconv.Atoi(idStr); err == nil {
					var existing models.Product
					if err := db.Preload("Categories").First(&existing, id).Error; err == nil {
						existing.EName = product.EName
						existing.ARName = product.ARName
						existing.EDescription = product.EDescription
						existing.ARDescription = product.ARDescription
						existing.SalePrice = product.SalePrice
						existing.RegularPrice = product.RegularPrice
						existing.BaseCost = product.BaseCost
						existing.Weight = product.Weight
						existing.Stock = product.Stock
						existing.Image = product.Image

						// Replace categories
						if err := db.Model(&existing).Association("Categories").Replace(categories); err != nil {
							skippedCount++
							continue
						}

						if err := db.Save(&existing).Error; err == nil {
							updatedCount++
							continue
						}
					}
				}
			}

			// Insert new product
			if err := db.Create(&product).Error; err == nil {
				createdCount++
			} else {
				skippedCount++
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message":       "Import completed",
			"created_count": createdCount,
			"updated_count": updatedCount,
			"skipped_count": skippedCount,
		})
	}
}
