package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"gorm.io/gorm"
)

type UpdateUserInput struct {
	Name    *string         `json:"name"`
	Phone   *string         `json:"phone"`
	Picture *string         `json:"picture"`
	Address *models.Address `json:"address"`
}

type CartItemInput struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

// GET /user
func GetUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		var user models.User

		if err := db.Preload("CartItems").First(&user, "id = ?", userID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		c.JSON(http.StatusOK, user)
	}
}

// GET /users
func GetAllUsers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var users []models.User
		if err := db.
			Select("id", "email", "name", "picture", "provider", "created_at"). // Select only public fields
			Order("created_at desc").
			Find(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
			return
		}

		c.JSON(http.StatusOK, users)
	}
}

// PUT /user
func UpdateUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		var user models.User

		if err := db.First(&user, "id = ?", userID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}

		var input UpdateUserInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		updates := make(map[string]interface{})
		if input.Name != nil {
			updates["name"] = *input.Name
		}
		if input.Phone != nil {
			updates["phone"] = *input.Phone
		}
		if input.Picture != nil {
			updates["picture"] = *input.Picture
		}
		if input.Address != nil {
			updates["street"] = input.Address.Street
			updates["city"] = input.Address.City
			updates["state"] = input.Address.State
			updates["postal_code"] = input.Address.PostalCode
			updates["country"] = input.Address.Country
		}

		if len(updates) > 0 {
			if err := db.Model(&user).Updates(updates).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
				return
			}
		}

		c.JSON(http.StatusOK, user)
	}
}

// GET /user/cart
func GetUserCart(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		var cart []models.CartItem

		if err := db.Where("user_id = ?", userID).Order("added_at desc").Find(&cart).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch cart"})
			return
		}

		c.JSON(http.StatusOK, cart)
	}
}

// POST /user/cart
func UpdateCartItem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")

		var input CartItemInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var item models.CartItem
		err := db.Where("user_id = ? AND product_id = ?", userID, input.ProductID).First(&item).Error

		if err == gorm.ErrRecordNotFound {
			item = models.CartItem{
				UserID:    userID.(string),
				ProductID: input.ProductID,
				Quantity:  input.Quantity,
				AddedAt:   time.Now(),
			}
			if err := db.Create(&item).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item"})
				return
			}
			c.JSON(http.StatusCreated, item)
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
			return
		}

		// Update quantity
		item.Quantity = input.Quantity
		if err := db.Save(&item).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item"})
			return
		}
		c.JSON(http.StatusOK, item)
	}
}

// DELETE /user/cart/:product_id
func DeleteCartItem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		productID := c.Param("product_id")

		if err := db.Where("user_id = ? AND product_id = ?", userID, productID).
			Delete(&models.CartItem{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Cart item deleted"})
	}
}
