package orderControllers

import (
	"errors"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// -------- Request Structs --------
type PlaceOrderRequest struct {
	UserID        string `json:"user_id" binding:"required"`
	Status        string `json:"status" binding:"required"`         // e.g. "pending", "shipped"
	PaymentStatus string `json:"payment_status" binding:"required"` // e.g. "paid", "failed"
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type UpdatePaymentStatusRequest struct {
	PaymentStatus string `json:"payment_status" binding:"required"`
}

// -------- Helpers --------

// Map string to OrderStatus
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

// Map string to PaymentStatus
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

// Generate unique order reference
func generateOrderRef(userID string) string {
	// Example: 20250908130500-<uuid4>
	return time.Now().Format("20060102150405") + "-" + uuid.NewString()
}

// -------- Core Logic --------

// PlaceOrder creates a new order for a user
func PlaceOrder(db *gorm.DB, req PlaceOrderRequest) error {
	var cart models.Cart
	if err := db.Preload("Items").Where("user_id = ?", req.UserID).First(&cart).Error; err != nil {
		return err
	}
	if len(cart.Items) == 0 {
		return errors.New("cart is empty")
	}

	orderStatus, err := mapOrderStatus(req.Status)
	if err != nil {
		return err
	}
	paymentStatus, err := mapPaymentStatus(req.PaymentStatus)
	if err != nil {
		return err
	}

	var total, totalWeight float64
	var orderItems []models.OrderItem

	return db.Transaction(func(tx *gorm.DB) error {
		// Process cart items
		for _, item := range cart.Items {
			var product models.Product
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&product, "id = ?", item.ProductID).Error; err != nil {
				return err
			}

			if product.Stock < item.Quantity {
				return errors.New("insufficient stock for product: " + item.ProductEName)
			}

			// Deduct stock
			product.Stock -= item.Quantity
			if err := tx.Save(&product).Error; err != nil {
				return err
			}

			// Accumulate totals
			total += item.ProductSalePrice * float64(item.Quantity)
			totalWeight += item.Weight * float64(item.Quantity)

			// Append to order items
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

		// Shipping calculation
		shippingCost := 0.0
		if totalWeight > 0 {
			shippingCost = float64(int(math.Ceil((totalWeight-1)/30.0))) * 30.0
		}

		// Create order
		order := models.Order{
			UserID:        req.UserID,
			Items:         orderItems,
			TotalAmount:   total + shippingCost,
			ShippingCost:  shippingCost,
			Status:        orderStatus,
			PaymentStatus: paymentStatus,
			OrderRef:      generateOrderRef(req.UserID), // ✅ unique ref
			CreatedAt:     time.Now(),
		}

		if err := tx.Create(&order).Error; err != nil {
			return err
		}

		// Clear cart items
		if err := tx.Where("cart_id = ?", cart.CartID).Delete(&models.CartItem{}).Error; err != nil {
			return err
		}

		return nil
	})
}

// -------- Handlers --------

// Place order (user)
func PlaceOrderHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req PlaceOrderRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := PlaceOrder(db, req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Order placed successfully"})
	}
}

// orderControllers (handlers portion)

func GetAllOrdersHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var orders []models.Order
		if err := db.
			Preload("User").
			Preload("Items").
			Preload("Items.Product").
			Order("created_at DESC").
			Find(&orders).Error; err != nil {
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
		var orders []models.Order
		if err := db.
			Where("user_id = ?", userID).
			Preload("User").
			Preload("Items").
			Preload("Items.Product").
			Order("created_at DESC").
			Find(&orders).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, orders)
	}
}

// New: Get single order by ID or order_ref
func GetOrderByIDHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("orderID") // you can pass numeric id or order_ref depending on route
		if id == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "orderID is required"})
			return
		}

		var order models.Order
		// Try numeric id first; if not numeric, fallback to order_ref lookup
		if err := db.
			Preload("User").
			Preload("Items").
			Preload("Items.Product").
			Where("id = ? OR order_ref = ?", id, id).
			First(&order).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, order)
	}
}

// Update order status
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
		if err := db.Model(&models.Order{}).Where("id = ?", orderID).
			Update("status", newStatus).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update order status"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Order status updated successfully"})
	}
}

// Update payment status
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
		newStatus, err := mapPaymentStatus(req.PaymentStatus)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := db.Model(&models.Order{}).Where("id = ?", orderID).
			Update("payment_status", newStatus).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update payment status"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Payment status updated successfully"})
	}
}

// Delete order
func DeleteOrderHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("orderID")
		if orderID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "orderID is required"})
			return
		}
		err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Where("order_id = ?", orderID).
				Delete(&models.OrderItem{}).Error; err != nil {
				return err
			}
			if err := tx.Where("id = ?", orderID).
				Delete(&models.Order{}).Error; err != nil {
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
