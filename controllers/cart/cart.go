package cartControllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"gorm.io/gorm"
)

type CartItemInput struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

// POST /user/cart
func UpdateCartItem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		userID := userIDVal.(string)

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

		// Check if user has a cart
		var cart models.Cart
		if err := db.Where("user_id = ?", userID).First(&cart).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User cart not found"})
			return
		}

		// Check if item already exists in the cart
		var item models.CartItem
		err := db.Where("cart_id = ? AND product_id = ?", cart.CartID, input.ProductID).First(&item).Error
		if err != nil {
			// New cart item
			if err == gorm.ErrRecordNotFound {
				newItem := models.CartItem{
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
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item to cart"})
					return
				}
				c.JSON(http.StatusCreated, newItem)
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cart item"})
			return
		}

		// Update existing cart item quantity and time
		item.Quantity = input.Quantity
		item.AddedAt = time.Now()
		if err := db.Save(&item).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart item"})
			return
		}

		c.JSON(http.StatusOK, item)
	}
}

// DELETE /user/cart/:product_id
func DeleteCartItem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context
		userIDVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		userID := userIDVal.(string)
		productID := c.Param("product_id")

		// Get the user's cart
		var cart models.Cart
		if err := db.Where("user_id = ?", userID).First(&cart).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User cart not found"})
			return
		}

		// Attempt to delete the cart item
		result := db.Where("cart_id = ? AND product_id = ?", cart.CartID, productID).Delete(&models.CartItem{})
		if result.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item"})
			return
		}

		// Check if item was actually deleted
		if result.RowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Cart item not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Cart item deleted"})
	}
}

// DELETE /user/cart
func ClearUserCart(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		userID := userIDVal.(string)

		var cart models.Cart
		if err := db.Where("user_id = ?", userID).First(&cart).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user cart"})
			return
		}

		if err := db.Where("cart_id = ?", cart.CartID).Delete(&models.CartItem{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear cart"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Cart cleared"})
	}
}

// GET /user/cart
func GetUserCart(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		userID := userIDVal.(string)

		var cart models.Cart
		if err := db.Preload("Items").Where("user_id = ?", userID).First(&cart).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cart"})
			return
		}

		c.JSON(http.StatusOK, cart.Items)
	}
}

// GET /user/cart
func GetAdminUserCart(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("user_id")

		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
			return
		}

		var cart models.Cart
		if err := db.Preload("Items").Where("user_id = ?", userID).First(&cart).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cart"})
			return
		}

		c.JSON(http.StatusOK, cart.Items)
	}
}
