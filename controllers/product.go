package controllers

import (
	"context"
	"fmt"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"github.com/junaidrashid-git/ecommerce-api/utils"
	"gorm.io/gorm"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// CreateProduct handles the creation of a new product with multiple categories.

func CreateProduct(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Read form fields
		ename := c.PostForm("ename")
		arname := c.PostForm("arname")
		edescription := c.PostForm("edescription")
		ardescription := c.PostForm("ardescription")
		salePriceStr := c.PostForm("sale_price")
		regularPriceStr := c.PostForm("regular_price")
		baseCostStr := c.PostForm("base_cost")
		weightStr := c.PostForm("weight")
		categoryIDsStr := c.PostForm("category_ids")

		// Required field check
		if ename == "" || salePriceStr == "" || weightStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "ename, sale_price, and weight are required"})
			return
		}

		// Parse numeric fields
		salePrice, err := strconv.ParseFloat(salePriceStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sale price"})
			return
		}

		weight, err := strconv.ParseFloat(weightStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid weight"})
			return
		}

		var regularPrice, baseCost float64
		if regularPriceStr != "" {
			regularPrice, err = strconv.ParseFloat(regularPriceStr, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid regular price"})
				return
			}
		}

		if baseCostStr != "" {
			baseCost, err = strconv.ParseFloat(baseCostStr, 64)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid base cost"})
				return
			}
		}

		// Parse category IDs
		var categories []models.Category
		if categoryIDsStr != "" {
			var categoryIDs []uint
			for _, idStr := range strings.Split(categoryIDsStr, ",") {
				id, err := strconv.ParseUint(strings.TrimSpace(idStr), 10, 64)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID: " + idStr})
					return
				}
				categoryIDs = append(categoryIDs, uint(id))
			}

			if err := db.Where("id IN ?", categoryIDs).Find(&categories).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
				return
			}
		}

		// Handle image upload to Cloudinary
		file, err := c.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Image is required"})
			return
		}

		// Open the uploaded file
		fileReader, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open image file"})
			return
		}
		defer fileReader.Close()

		// Initialize Cloudinary
		cld := utils.InitCloudinary()

		// Upload to Cloudinary
		uploadResult, err := cld.Upload.Upload(context.Background(), fileReader, uploader.UploadParams{
			Folder:   "ecommerce/products", // Optional: Folder path in Cloudinary
			PublicID: fmt.Sprintf("%d_%s", time.Now().Unix(), strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename))),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload image to Cloudinary"})
			return
		}

		// Use the secure URL from Cloudinary
		imageURL := uploadResult.SecureURL

		// Create product record
		product := models.Product{
			EName:         ename,
			ARName:        arname,
			EDescription:  edescription,
			ARDescription: ardescription,
			SalePrice:     salePrice,
			RegularPrice:  regularPrice,
			BaseCost:      baseCost,
			Weight:        weight,
			Image:         imageURL, // Save Cloudinary URL
			Categories:    categories,
		}

		if err := db.Create(&product).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
			return
		}

		c.JSON(http.StatusCreated, product)
	}
}

func UpdateProduct(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var product models.Product

		// Fetch the product by ID
		if err := db.Preload("Categories").First(&product, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		// Read form fields (PUT with multipart/form-data)
		ename := c.PostForm("ename")
		arname := c.PostForm("arname")
		edescription := c.PostForm("edescription")
		ardescription := c.PostForm("ardescription")
		salePriceStr := c.PostForm("sale_price")
		regularPriceStr := c.PostForm("regular_price")
		baseCostStr := c.PostForm("base_cost")
		weightStr := c.PostForm("weight")
		categoryIDsStr := c.PostForm("category_ids")

		// Update simple fields if provided
		if ename != "" {
			product.EName = ename
		}
		if arname != "" {
			product.ARName = arname
		}
		if edescription != "" {
			product.EDescription = edescription
		}
		if ardescription != "" {
			product.ARDescription = ardescription
		}
		if salePriceStr != "" {
			if salePrice, err := strconv.ParseFloat(salePriceStr, 64); err == nil {
				product.SalePrice = salePrice
			}
		}
		if regularPriceStr != "" {
			if regularPrice, err := strconv.ParseFloat(regularPriceStr, 64); err == nil {
				product.RegularPrice = regularPrice
			}
		}
		if baseCostStr != "" {
			if baseCost, err := strconv.ParseFloat(baseCostStr, 64); err == nil {
				product.BaseCost = baseCost
			}
		}
		if weightStr != "" {
			if weight, err := strconv.ParseFloat(weightStr, 64); err == nil {
				product.Weight = weight
			}
		}

		// Update categories if provided
		if categoryIDsStr != "" {
			var categoryIDs []uint
			for _, idStr := range strings.Split(categoryIDsStr, ",") {
				id, err := strconv.ParseUint(strings.TrimSpace(idStr), 10, 64)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID: " + idStr})
					return
				}
				categoryIDs = append(categoryIDs, uint(id))
			}

			var categories []models.Category
			if err := db.Where("id IN ?", categoryIDs).Find(&categories).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
				return
			}
			product.Categories = categories
		}

		// Handle new image upload if provided
		file, err := c.FormFile("image")
		if err == nil {
			fileReader, err := file.Open()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open image file"})
				return
			}
			defer fileReader.Close()

			cld := utils.InitCloudinary()

			// Optionally: delete old image from Cloudinary
			if product.Image != "" {
				// Get public ID from old image URL
				parts := strings.Split(product.Image, "/")
				publicID := strings.TrimSuffix(parts[len(parts)-1], filepath.Ext(parts[len(parts)-1]))
				cld.Upload.Destroy(context.Background(), uploader.DestroyParams{PublicID: "ecommerce/products/" + publicID})
			}

			// Upload new image
			uploadResult, err := cld.Upload.Upload(context.Background(), fileReader, uploader.UploadParams{
				Folder:   "ecommerce/products",
				PublicID: fmt.Sprintf("%d_%s", time.Now().Unix(), strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename))),
			})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload new image to Cloudinary"})
				return
			}
			product.Image = uploadResult.SecureURL
		} else if err != http.ErrMissingFile {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid image upload"})
			return
		}

		if err := db.Session(&gorm.Session{FullSaveAssociations: true}).Save(&product).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
			return
		}

		c.JSON(http.StatusOK, product)
	}
}

// DeleteProduct handles deleting a product by ID.
func DeleteProduct(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var product models.Product

		if err := db.First(&product, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		if err := db.Delete(&product).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
	}
}

// GetProducts fetches all products including their categories.
// GetProducts handles filtered product listing
func GetProducts(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var products []models.Product

		search := c.Query("search")
		categoryID := c.Query("category_id")
		brandID := c.Query("brand_id")

		query := db.Preload("Categories")

		// Filter by product name or description (case-insensitive partial match)
		if search != "" {
			searchPattern := "%" + search + "%"
			query = query.Where("e_name ILIKE ? OR ar_name ILIKE ? OR e_description ILIKE ?", searchPattern, searchPattern, searchPattern)
		}

		// Filter by category ID
		if categoryID != "" {
			query = query.Joins("JOIN product_categories ON product_categories.product_id = products.id").
				Where("product_categories.category_id = ?", categoryID)
		}

		// Filter by brand ID (if your Product model has a BrandID field)
		if brandID != "" {
			query = query.Where("brand_id = ?", brandID)
		}

		// Execute query
		if err := query.Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
			return
		}

		c.JSON(http.StatusOK, products)
	}
}
