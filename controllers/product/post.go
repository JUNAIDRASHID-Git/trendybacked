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

// CreateProduct creates a new product with multiple categories + image upload.
// Expects multipart/form-data with fields:
//   - ename (required), sale_price (required), weight (required)
//   - arname, edescription, ardescription, regular_price (optional), base_cost (optional)
//   - category_ids (comma-separated, optional), and an "image" file (required).
func CreateProduct(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1️⃣ Parse mandatory form fields
		ename := c.PostForm("ename")
		salePriceStr := c.PostForm("sale_price")
		weightStr := c.PostForm("weight")
		if ename == "" || salePriceStr == "" || weightStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ename, sale_price, and weight are required"})
			return
		}

		// 2️⃣ Parse optional fields
		arname := c.PostForm("arname")
		edescription := c.PostForm("edescription")
		ardescription := c.PostForm("ardescription")
		regularPriceStr := c.PostForm("regular_price")
		baseCostStr := c.PostForm("base_cost")
		categoryIDsStr := c.PostForm("category_ids")

		// 3️⃣ Convert numeric fields; detailed validations
		salePrice, err := strconv.ParseFloat(salePriceStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sale_price"})
			return
		}
		weight, err := strconv.ParseFloat(weightStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid weight"})
			return
		}

		var regularPrice, baseCost float64
		if regularPriceStr != "" {
			if rp, parseErr := strconv.ParseFloat(regularPriceStr, 64); parseErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid regular_price"})
				return
			} else {
				regularPrice = rp
			}
		}
		if baseCostStr != "" {
			if bc, parseErr := strconv.ParseFloat(baseCostStr, 64); parseErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid base_cost"})
				return
			} else {
				baseCost = bc
			}
		}

		// 4️⃣ Parse category IDs (if provided)
		var categories []models.Category
		if categoryIDsStr != "" {
			idTokens := strings.Split(categoryIDsStr, ",")
			var parsedIDs []uint
			for _, tok := range idTokens {
				tok = strings.TrimSpace(tok)
				if tok == "" {
					continue
				}
				if id64, parseErr := strconv.ParseUint(tok, 10, 64); parseErr != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category_ids format"})
					return
				} else {
					parsedIDs = append(parsedIDs, uint(id64))
				}
			}
			// Fetch actual Category records
			if len(parsedIDs) > 0 {
				if err := db.Where("id IN ?", parsedIDs).Find(&categories).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
					return
				}
			}
		}

		// 5️⃣ Handle image upload to Cloudinary
		file, err := c.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Image is required"})
			return
		}
		fileReader, openErr := file.Open()
		if openErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open image file"})
			return
		}
		defer fileReader.Close()

		cld := utils.InitCloudinary()
		uploadParams := uploader.UploadParams{
			Folder: "ecommerce/products",
			// PublicID: timestamp_filenameWithoutExtension
			PublicID: fmt.Sprintf("%d_%s", time.Now().Unix(), strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename))),
		}
		uploadResult, uploadErr := cld.Upload.Upload(context.Background(), fileReader, uploadParams)
		if uploadErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image to Cloudinary"})
			return
		}
		imageURL := uploadResult.SecureURL

		// 6️⃣ Begin transaction: create product + associations atomically
		tx := db.Begin()
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start database transaction"})
			return
		}

		// 7️⃣ Build Product struct
		newProduct := models.Product{
			EName:         ename,
			ARName:        arname,
			EDescription:  edescription,
			ARDescription: ardescription,
			SalePrice:     salePrice,
			RegularPrice:  regularPrice,
			BaseCost:      baseCost,
			Weight:        weight,
			Image:         imageURL,
			Categories:    categories, // GORM will insert into join table
		}

		// 8️⃣ Create Product + set up join‐table entries
		if err := tx.Create(&newProduct).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
			return
		}

		// 9️⃣ Commit transaction
		if err := tx.Commit().Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit product creation"})
			return
		}

		// 1️⃣0️⃣ Return created product (with Categories populated)
		c.JSON(http.StatusCreated, newProduct)
	}
}