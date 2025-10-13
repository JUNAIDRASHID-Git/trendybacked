package orderControllers

import (
	"errors"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Struct to receive client request
type PlaceOrderRequest struct {
	CartID        string `json:"cart_id" binding:"required"`
	Status        string `json:"status" binding:"required"`         // e.g. "pending", "shipped"
	PaymentStatus string `json:"payment_status" binding:"required"` // e.g. "paid", "failed"
}

// Utility: map and validate status
func mapOrderStatus(status string) (models.OrderStatus, error) {
	switch strings.ToLower(status) {
	case string(models.OrderStatusPending):
		return models.OrderStatusPending, nil
	case string(models.OrderStatusConfirmed):
		return models.OrderStatusConfirmed, nil
	case string(models.OrderStatusReadyToShip):
		return models.OrderStatusReadyToShip, nil
	case string(models.OrderStatusShipped):
		return models.OrderStatusShipped, nil
	case string(models.OrderStatusDelivered):
		return models.OrderStatusDelivered, nil
	case string(models.OrderStatusReturned):
		return models.OrderStatusReturned, nil
	case string(models.OrderStatusCancelled):
		return models.OrderStatusCancelled, nil
	default:
		return "", errors.New("invalid order status")
	}
}

// Utility: map and validate payment status
func mapPaymentStatus(status string) (models.PaymentStatus, error) {
	switch strings.ToLower(status) {
	case string(models.PaymentStatusPending):
		return models.PaymentStatusPending, nil
	case string(models.PaymentStatusPaid):
		return models.PaymentStatusPaid, nil
	case string(models.PaymentStatusFailed):
		return models.PaymentStatusFailed, nil
	case string(models.PaymentStatusRefunded):
		return models.PaymentStatusRefunded, nil
	default:
		return "", errors.New("invalid payment status")
	}
}

// Place order from a given CartID (used for webhook or API)
func PlaceOrder(db *gorm.DB, cartID, status, paymentStatus string) error {
	var cart models.Cart
	err := db.Preload("Items").Where("cart_id = ?", cartID).First(&cart).Error
	if err != nil {
		return errors.New("cart not found for cartID: " + cartID)
	}
	if len(cart.Items) == 0 {
		return errors.New("cart is empty")
	}

	mappedOrderStatus, _ := mapOrderStatus(status)
	mappedPaymentStatus, _ := mapPaymentStatus(paymentStatus)

	var total, totalWeight float64
	var orderItems []models.OrderItem

	return db.Transaction(func(tx *gorm.DB) error {
		for _, item := range cart.Items {
			var product models.Product
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, "id = ?", item.ProductID).Error; err != nil {
				return err
			}

			if product.Stock < item.Quantity {
				return errors.New("insufficient stock for product: " + product.EName)
			}

			product.Stock -= item.Quantity
			if err := tx.Save(&product).Error; err != nil {
				return err
			}

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

		// === Shipping cost pattern (same as your Flutter logic) ===
		var shippingCost float64
		if totalWeight <= 0 {
			shippingCost = 0.0
		} else if totalWeight <= 29 {
			shippingCost = 30.0
		} else if totalWeight <= 59 {
			shippingCost = 60.0
		} else if totalWeight <= 89 {
			shippingCost = 90.0
		} else {
			extraBlocks := int(math.Ceil((totalWeight - 89) / 30.0))
			shippingCost = 90.0 + float64(extraBlocks*30)
		}

		totalWithShipping := total + shippingCost

		order := models.Order{
			UserID:        cart.UserID,
			Items:         orderItems,
			TotalAmount:   totalWithShipping,
			ShippingCost:  shippingCost,
			Status:        mappedOrderStatus,
			PaymentStatus: mappedPaymentStatus,
			CreatedAt:     time.Now(),
		}

		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		if err := tx.Where("cart_id = ?", cart.CartID).Delete(&models.CartItem{}).Error; err != nil {
			return err
		}

		go BroadcastNewOrder(order)
		return nil
	})
}

// HTTP handler to place order
func PlaceOrderHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req PlaceOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := PlaceOrder(db, req.CartID, req.Status, req.PaymentStatus); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Order placed successfully"})
	}
}

// Handler for fetching all orders (Admin)
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

// Handler for fetching a user's orders
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

// Fetch all orders with related user and items
func GetAllOrders(db *gorm.DB) ([]models.Order, error) {
	var orders []models.Order
	err := db.Preload("Items").Preload("User").Order("created_at DESC").Find(&orders).Error
	return orders, err
}

// Fetch user-specific orders
func GetUserOrders(db *gorm.DB, userID string) ([]models.Order, error) {
	var orders []models.Order
	err := db.Where("user_id = ?", userID).Preload("Items").Order("created_at DESC").Find(&orders).Error
	return orders, err
}

// Request struct to update order status
type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"` // e.g. "shipped", "cancelled"
}

// Handler to update order status
func UpdateOrderStatusHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("orderID")
		if orderID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "orderID is required"})
			return
		}

		var req UpdateOrderStatusRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		newStatus, err := mapOrderStatus(req.Status)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := db.Model(&models.Order{}).Where("id = ?", orderID).Update("status", newStatus).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update order status"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Order status updated successfully"})
	}
}

// Request struct to update payment status
type UpdatePaymentStatusRequest struct {
	PaymentStatus string `json:"payment_status" binding:"required"` // e.g. "paid", "failed", "refunded"
}

// Handler to update the payment status of an order
func UpdatePaymentStatusHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("orderID")
		if orderID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "orderID is required"})
			return
		}

		var req UpdatePaymentStatusRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate and map payment status
		newStatus, err := mapPaymentStatus(req.PaymentStatus)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Update payment_status field
		if err := db.Model(&models.Order{}).Where("id = ?", orderID).Update("payment_status", newStatus).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update payment status"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Payment status updated successfully"})
	}
}

// Handler to delete an order and its items
func DeleteOrderHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("orderID")
		if orderID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "orderID is required"})
			return
		}

		err := db.Transaction(func(tx *gorm.DB) error {
			// Delete order items first (CASCADE should also handle this, but do it explicitly for safety)
			if err := tx.Where("order_id = ?", orderID).Delete(&models.OrderItem{}).Error; err != nil {
				return err
			}

			// Delete order itself
			if err := tx.Where("id = ?", orderID).Delete(&models.Order{}).Error; err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete order"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Order deleted successfully"})
	}
}
