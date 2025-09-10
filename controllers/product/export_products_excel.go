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

func ExportProductsToExcel(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var products []models.Product
		if err := db.Preload("Categories").Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
			return
		}

		file := xlsx.NewFile()
		sheet, err := file.AddSheet("Products")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create Excel sheet"})
			return
		}

		// Header row
		headers := []string{
			"ID", "EName", "ARName", "EDescription", "ARDescription",
			"SalePrice", "RegularPrice", "BaseCost", "Weight", "Stock",
			"Image", "CategoryIDs", "CreatedAt", "UpdatedAt",
		}
		headerRow := sheet.AddRow()
		for _, h := range headers {
			headerRow.AddCell().SetValue(h)
		}

		// Data rows
		for _, p := range products {
			row := sheet.AddRow()

			row.AddCell().SetValue(p.ID)
			row.AddCell().SetValue(p.EName)
			row.AddCell().SetValue(p.ARName)
			row.AddCell().SetValue(p.EDescription)
			row.AddCell().SetValue(p.ARDescription)
			row.AddCell().SetValue(p.SalePrice)
			row.AddCell().SetValue(p.RegularPrice)
			row.AddCell().SetValue(p.BaseCost)
			row.AddCell().SetValue(p.Weight)
			row.AddCell().SetValue(p.Stock)
			row.AddCell().SetValue(p.Image)

			var catIDs []string
			for _, cat := range p.Categories {
				catIDs = append(catIDs, strconv.Itoa(int(cat.ID)))
			}
			row.AddCell().SetValue(strings.Join(catIDs, ","))

			row.AddCell().SetValue(p.CreatedAt.Format("2006-01-02 15:04:05"))
			row.AddCell().SetValue(p.UpdatedAt.Format("2006-01-02 15:04:05"))
		}

		// Set response headers for download
		c.Header("Content-Disposition", "attachment; filename=products.xlsx")
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Transfer-Encoding", "binary")
		c.Header("Expires", "0")

		// Write file to response
		if err := file.Write(c.Writer); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write Excel file"})
			return
		}
	}
}
