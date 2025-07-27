package productcontroller

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"github.com/junaidrashid-git/ecommerce-api/utils"
	"gorm.io/gorm"
)

// UpdateProduct updates an existing product’s fields, categories, and optionally its image.
// Expects multipart/form-data with any of:
//   - ename, arname, edescription, ardescription, sale_price, regular_price, base_cost, weight
//   - category_ids (comma-separated) to fully replace associations
//   - image (to replace existing image)
//
// Returns 200 + updated product, or appropriate error.
func UpdateProduct(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1️⃣ Parse product ID from URL
		id := c.Param("id")
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Product ID is required"})
			return
		}

		// 2️⃣ Fetch existing product with categories
		var existing models.Product
		if err := db.Preload("Categories").First(&existing, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		// 3️⃣ Parse form fields
		ename := c.PostForm("ename")
		arname := c.PostForm("arname")
		edescription := c.PostForm("edescription")
		ardescription := c.PostForm("ardescription")
		salePriceStr := c.PostForm("sale_price")
		regularPriceStr := c.PostForm("regular_price")
		baseCostStr := c.PostForm("base_cost")
		weightStr := c.PostForm("weight")
		stockStr := c.PostForm("stock")
		categoryIDsStr := c.PostForm("category_ids")

		// 4️⃣ Begin transaction for update + associations
		tx := db.Begin()
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
			return
		}

		// 5️⃣ Update scalar fields if provided
		if ename != "" {
			existing.EName = ename
		}
		if arname != "" {
			existing.ARName = arname
		}
		if edescription != "" {
			existing.EDescription = edescription
		}
		if ardescription != "" {
			existing.ARDescription = ardescription
		}
		if salePriceStr != "" {
			if sp, err := strconv.ParseFloat(salePriceStr, 64); err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sale_price format"})
				return
			} else {
				existing.SalePrice = sp
			}
		}
		if regularPriceStr != "" {
			if rp, err := strconv.ParseFloat(regularPriceStr, 64); err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid regular_price format"})
				return
			} else {
				existing.RegularPrice = rp
			}
		}
		if baseCostStr != "" {
			if bc, err := strconv.ParseFloat(baseCostStr, 64); err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid base_cost format"})
				return
			} else {
				existing.BaseCost = bc
			}
		}
		if weightStr != "" {
			if w, err := strconv.ParseFloat(weightStr, 64); err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid weight format"})
				return
			} else {
				existing.Weight = w
			}
		}

		if stockStr != "" {
			if stock, err := strconv.Atoi(stockStr); err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid stock format"})
				return
			} else {
				existing.Stock = stock
			}
		}

		// 6️⃣ Handle category associations if provided
		if categoryIDsStr != "" {
			idTokens := strings.Split(categoryIDsStr, ",")
			var parsedIDs []uint
			for _, tok := range idTokens {
				tok = strings.TrimSpace(tok)
				if tok == "" {
					continue
				}
				if id64, err := strconv.ParseUint(tok, 10, 64); err != nil {
					tx.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category_ids format"})
					return
				} else {
					parsedIDs = append(parsedIDs, uint(id64))
				}
			}

			// Fetch categories that exist
			var newCategories []models.Category
			if len(parsedIDs) > 0 {
				if err := tx.Where("id IN ?", parsedIDs).Find(&newCategories).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
					return
				}
			}

			// Replace association in join table
			if err := tx.Model(&existing).Association("Categories").Replace(newCategories); err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update category associations"})
				return
			}
			existing.Categories = newCategories
		}

		// 7️⃣ Handle optional image update
		file, fileErr := c.FormFile("image")
		if fileErr == nil {
			// Open new image
			fileReader, openErr := file.Open()
			if openErr != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open new image"})
				return
			}
			defer fileReader.Close()

			cld := utils.InitCloudinary()
			// Delete old image from Cloudinary (if exists)
			if existing.Image != "" {
				fragments := strings.Split(existing.Image, "/")
				publicID := strings.TrimSuffix(fragments[len(fragments)-1], filepath.Ext(fragments[len(fragments)-1]))
				_, _ = cld.Upload.Destroy(context.Background(), uploader.DestroyParams{
					PublicID: "ecommerce/products/" + publicID,
				})
			}

			// Upload new image
			uploadParams := uploader.UploadParams{
				Folder:   "ecommerce/products",
				PublicID: fmt.Sprintf("%d_%s", time.Now().Unix(), strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename))),
			}
			uploadResult, uploadErr := cld.Upload.Upload(context.Background(), fileReader, uploadParams)
			if uploadErr != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload new image"})
				return
			}
			existing.Image = uploadResult.SecureURL
		} else if fileErr != http.ErrMissingFile {
			// If error is not “no file provided,” treat as a genuine error
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image upload"})
			return
		}

		// 8️⃣ Save all changes to product
		if err := tx.Save(&existing).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
			return
		}

		// 9️⃣ Commit transaction
		if err := tx.Commit().Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit product update"})
			return
		}

		// 1️⃣0️⃣ Return updated product (with Categories)
		c.JSON(http.StatusOK, existing)
	}
}
