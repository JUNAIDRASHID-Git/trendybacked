package routes

import (
	"github.com/gin-gonic/gin"
	orderControllers "github.com/junaidrashid-git/ecommerce-api/controllers/order"
	"gorm.io/gorm"
)

func SetupOrderRoutes(r *gin.Engine, db *gorm.DB) {
	orders := r.Group("/orders")
	{
		// Create a new order
		orders.POST("/place", orderControllers.PlaceOrderHandler(db))

		// Fetch all orders (admin)
		orders.GET("/", orderControllers.GetAllOrdersHandler(db))

		// websocket endpoint for real-time order updates
		orders.GET("/ws/orders", orderControllers.OrderWebSocketHandler)

		// Fetch orders for a specific user
		orders.GET("/user/:userID", orderControllers.GetUserOrdersHandler(db))

		// Update order status (e.g., shipped, cancelled)
		orders.PUT("/:orderID/status", orderControllers.UpdateOrderStatusHandler(db))

		// Update payment status (e.g., paid, refunded)
		orders.PUT("/:orderID/payment-status", orderControllers.UpdatePaymentStatusHandler(db))

		// Delete an order
		orders.DELETE("/:orderID", orderControllers.DeleteOrderHandler(db))
	}
}
