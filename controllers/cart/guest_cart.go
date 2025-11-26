package cartControllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"gorm.io/gorm"
)

// POST /guest/cart
func UpdateGuestCartItem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		guestID := c.Query("guest_id")
		if guestID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "guest_id is required"})
			return
		}

		var input CartItemInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
			return
		}

		// Fetch product from DB
		var product models.Product
		if err := db.First(&product, "id = ?", input.ProductID).Error; err != nil {
			status := http.StatusInternalServerError
			errMsg := "Failed to validate product"
			if err == gorm.ErrRecordNotFound {
				status = http.StatusBadRequest
				errMsg = "Product does not exist"
			}
			c.JSON(status, gin.H{"error": errMsg})
			return
		}

		// Check if guest has a cart
		var cart models.GuestCart
		if err := db.Where("guest_id = ?", guestID).First(&cart).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				cart = models.GuestCart{GuestID: guestID}
				if err := db.Create(&cart).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create guest cart"})
					return
				}
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch guest cart"})
				return
			}
		}

		// Check if item already exists
		var item models.GuestCartItem
		err := db.Where("cart_id = ? AND product_id = ?", cart.CartID, input.ProductID).First(&item).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				newItem := models.GuestCartItem{
					CartID:              cart.CartID,
					ProductID:           product.ID,
					ProductEName:        product.EName,
					ProductArName:       product.ARName,
					ProductImage:        product.Image,
					ProductStock:        product.Stock,
					ProductSalePrice:    product.SalePrice,
					ProductRegularPrice: product.RegularPrice,
					Weight:              product.Weight,
					Quantity:            input.Quantity,
					AddedAt:             time.Now(),
				}
				if err := db.Create(&newItem).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item to guest cart"})
					return
				}
				c.JSON(http.StatusCreated, newItem)
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cart item"})
			return
		}

		// Update existing item
		item.Quantity = input.Quantity
		item.AddedAt = time.Now()
		if err := db.Save(&item).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update guest cart item"})
			return
		}

		c.JSON(http.StatusOK, item)
	}
}

// DELETE /guest/cart/:product_id
func DeleteGuestCartItem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		guestID := c.Query("guest_id")
		if guestID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "guest_id is required"})
			return
		}

		// Convert product_id param to uint
		productIDParam := c.Param("product_id")
		productIDUint, err := strconv.ParseUint(productIDParam, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product_id"})
			return
		}
		productID := uint(productIDUint)

		// Get guest cart
		var cart models.GuestCart
		if err := db.Where("guest_id = ?", guestID).First(&cart).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Guest cart not found"})
			return
		}

		// Delete item
		result := db.Where("cart_id = ? AND product_id = ?", cart.CartID, productID).Delete(&models.GuestCartItem{})
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item"})
			return
		}
		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cart item not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Guest cart item deleted"})
	}
}

// DELETE /guest/cart
func ClearGuestCart(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		guestID := c.Query("guest_id")
		if guestID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "guest_id is required"})
			return
		}

		var cart models.GuestCart
		if err := db.Where("guest_id = ?", guestID).First(&cart).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch guest cart"})
			return
		}

		if err := db.Where("cart_id = ?", cart.CartID).Delete(&models.GuestCartItem{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear guest cart"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Guest cart cleared"})
	}
}

// GET /guest/cart
func GetGuestCart(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		guestID := c.Query("guest_id")
		if guestID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "guest_id is required"})
			return
		}

		var cart models.GuestCart
		if err := db.Preload("Items").Where("guest_id = ?", guestID).First(&cart).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusOK, []models.GuestCartItem{})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch guest cart"})
			return
		}

		c.JSON(http.StatusOK, cart.Items)
	}
}
