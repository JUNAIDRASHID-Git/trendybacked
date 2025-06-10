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

		var cart models.Cart
		if err := db.Where("user_id = ?", userID).First(&cart).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				cart = models.Cart{UserID: userID, CreatedAt: time.Now(), UpdatedAt: time.Now()}
				if err := db.Create(&cart).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user cart"})
					return
				}
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user cart"})
				return
			}
		}

		var item models.CartItem
		err := db.Where("cart_id = ? AND product_id = ?", cart.CartID, input.ProductID).First(&item).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				newItem := models.CartItem{
					CartID:    cart.CartID,
					ProductID: input.ProductID,
					Quantity:  input.Quantity,
					AddedAt:   time.Now(),
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
		userIDVal, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		userID := userIDVal.(string)
		productID := c.Param("product_id")

		var cart models.Cart
		if err := db.Where("user_id = ?", userID).First(&cart).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user cart"})
			return
		}

		if err := db.Where("cart_id = ? AND product_id = ?", cart.CartID, productID).
			Delete(&models.CartItem{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item"})
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
