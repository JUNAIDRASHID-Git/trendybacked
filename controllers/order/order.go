package orderControllers

import (
	"errors"
	"math"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"gorm.io/gorm"
)

// Request struct to parse JSON body for placing order
type PlaceOrderRequest struct {
	UserID        string `json:"user_id" binding:"required"`
	PaymentStatus string `json:"payment_status" binding:"required"`
}

// Updated PlaceOrder function now takes *gorm.DB and userID/paymentStatus as params
func PlaceOrder(db *gorm.DB, userID string, paymentStatus string) error {
	var cart models.Cart
	err := db.Preload("Items").Where("user_id = ?", userID).First(&cart).Error
	if err != nil {
		return err
	}
	if len(cart.Items) == 0 {
		return errors.New("cart is empty")
	}

	var total float64
	var totalWeight float64
	var orderItems []models.OrderItem

	for _, item := range cart.Items {
		total += item.ProductSalePrice * float64(item.Quantity)
		totalWeight += item.Weight * float64(item.Quantity)
		orderItems = append(orderItems, models.OrderItem{
			ProductID:           item.ProductID,
			ProductEName:        item.ProductEName,
			ProductArName:       item.ProductArName,
			ProductImage:        item.ProductImage,
			ProductSalePrice:    item.ProductSalePrice,
			ProductRegularPrice: item.ProductRegularPrice,
			Weight:              item.Weight,
			Quantity:            item.Quantity,
		})
	}

	// Calculate shipping cost: ((ceil(totalWeight / 30)) * 30)
	shippingCost := 0.0
	if totalWeight > 0 {
		shippingCost = float64(int(math.Ceil(totalWeight/30.0))) * 30.0
	}

	totalWithShipping := total + shippingCost

	order := models.Order{
		UserID:       userID,
		Items:        orderItems,
		TotalAmount:  totalWithShipping,
		ShippingCost: shippingCost,         // Make sure your Order model has this field
		Status:       models.StatusSuccess, // Map your paymentStatus here if needed
		CreatedAt:    time.Now(),
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		if err := tx.Where("cart_id = ?", cart.CartID).Delete(&models.CartItem{}).Error; err != nil {
			return err
		}
		return nil
	})
}

// HTTP handler that extracts JSON body, calls PlaceOrder, and returns JSON response
func PlaceOrderHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req PlaceOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := PlaceOrder(db, req.UserID, req.PaymentStatus)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Order placed successfully"})
	}
}

func GetAllOrdersHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		orders, err := GetAllOrders(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, orders)
	}
}

func GetUserOrdersHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Param("userID")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "userID is required"})
			return
		}
		orders, err := GetUserOrders(db, userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, orders)
	}
}

// Modified GetAllOrders and GetUserOrders to accept db *gorm.DB
func GetAllOrders(db *gorm.DB) ([]models.Order, error) {
	var orders []models.Order
	err := db.Preload("Items").Preload("User").Order("created_at DESC").Find(&orders).Error
	return orders, err
}

func GetUserOrders(db *gorm.DB, userID string) ([]models.Order, error) {
	var orders []models.Order
	err := db.Where("user_id = ?", userID).Preload("Items").Order("created_at DESC").Find(&orders).Error
	return orders, err
}
