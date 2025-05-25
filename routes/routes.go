package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/auth"
	"github.com/junaidrashid-git/ecommerce-api/controllers"
	"github.com/junaidrashid-git/ecommerce-api/middleware"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	// Public routes (no auth)
	r.POST("/auth/google", auth.GoogleAuthHandler(db))

	// User routes with JWT middleware
	userGroup := r.Group("/user")
	userGroup.Use(middleware.ValidateToken)
	{
		userGroup.GET("/user", controllers.GetUser(db))
		userGroup.PUT("/user", controllers.UpdateUser(db))
		userGroup.GET("/cart", controllers.GetUserCart(db))
		userGroup.POST("/cart", controllers.UpdateCartItem(db))
		userGroup.GET("/products", controllers.GetProducts(db))
		userGroup.DELETE("/cart/:product_id", controllers.DeleteCartItem(db))
	}

	// Admin routes with API key middleware
	adminGroup := r.Group("/admin")
	adminGroup.Use(middleware.ValidateAPIKey)
	{
		adminGroup.GET("/users", controllers.GetAllUsers(db))
		adminGroup.POST("/products", controllers.CreateProduct(db))
		adminGroup.PUT("/products/:id", controllers.UpdateProduct(db))
		adminGroup.DELETE("/products/:id", controllers.DeleteProduct(db))
		adminGroup.GET("/products", controllers.GetProducts(db))
		adminGroup.POST("/products/import-excel", controllers.ImportProductsFromExcel(db))
	}
}
