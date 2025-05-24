package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"github.com/tealeg/xlsx"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

func ImportProductsFromExcel(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get Excel file
		excelFileHeader, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Excel file is required"})
			return
		}

		excelFile, err := excelFileHeader.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open Excel file"})
			return
		}
		defer excelFile.Close()

		// 2. Parse Excel file
		xlFile, err := xlsx.OpenReaderAt(excelFile, excelFileHeader.Size)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Excel file"})
			return
		}

		if len(xlFile.Sheets) == 0 || xlFile.Sheets[0].MaxRow == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Excel file is empty"})
			return
		}

		sheet := xlFile.Sheets[0]

		createdCount := 0
		updatedCount := 0
		skippedCount := 0

		// 3. Loop through rows
		for i := 1; i < sheet.MaxRow; i++ {
			row := sheet.Rows[i]
			if len(row.Cells) < 8 {
				skippedCount++
				continue // skip incomplete rows
			}

			idStr := row.Cells[0].String()
			ename := row.Cells[1].String()
			arname := row.Cells[2].String()
			endescription := row.Cells[3].String()
			ardescription := row.Cells[4].String()
			salePriceStr := row.Cells[5].String()
			regularPriceStr := row.Cells[6].String()
			baseCostStr := row.Cells[7].String()
			weightStr := row.Cells[8].String()

			// Required field validation
			if ename == "" || salePriceStr == "" || weightStr == "" {
				skippedCount++
				continue
			}

			// Parse numerical values
			salePrice, _ := strconv.ParseFloat(salePriceStr, 64)
			regularPrice, _ := strconv.ParseFloat(regularPriceStr, 64)
			baseCost, _ := strconv.ParseFloat(baseCostStr, 64)
			weight, _ := strconv.ParseFloat(weightStr, 64)

			product := models.Product{
				EName:         ename,
				ARName:        arname,
				EDescription:  endescription,
				ARDescription: ardescription,
				SalePrice:     salePrice,
				RegularPrice:  regularPrice,
				BaseCost:      baseCost,
				Weight:        weight,
			}

			// Update if ID is present
			if idStr != "" {
				if id, err := strconv.Atoi(idStr); err == nil {
					var existing models.Product
					if err := db.First(&existing, id).Error; err == nil {
						existing.EName = product.EName
						existing.ARName = product.ARName
						existing.EDescription = product.EDescription
						existing.ARDescription = product.ARDescription
						existing.SalePrice = product.SalePrice
						existing.RegularPrice = product.RegularPrice
						existing.BaseCost = product.BaseCost
						existing.Weight = product.Weight

						if err := db.Save(&existing).Error; err == nil {
							updatedCount++
							continue
						}
					}
				}
			}

			// Create if new
			if err := db.Create(&product).Error; err == nil {
				createdCount++
			}
		}

		// 4. Respond with summary
		c.JSON(http.StatusOK, gin.H{
			"message":       "Import completed",
			"created_count": createdCount,
			"updated_count": updatedCount,
			"skipped_count": skippedCount,
		})
	}
}
