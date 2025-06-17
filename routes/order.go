package routes

import (
	"github.com/gin-gonic/gin"
	orderControllers "github.com/junaidrashid-git/ecommerce-api/controllers/order"
	"gorm.io/gorm"
)

func SetupOrderRoutes(r *gin.Engine, db *gorm.DB) {
	orders := r.Group("/orders")
	{
		orders.POST("/place", orderControllers.PlaceOrderHandler(db))
		orders.GET("/", orderControllers.GetAllOrdersHandler(db))
		orders.GET("/user/:userID", orderControllers.GetUserOrdersHandler(db))
	}
}
