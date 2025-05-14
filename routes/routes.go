package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/auth"
	"github.com/junaidrashid-git/ecommerce-api/controllers"
	"github.com/junaidrashid-git/ecommerce-api/middleware"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	// Google Auth Route
	r.POST("/auth/google", auth.GoogleAuthHandler(db))
	// Protect /users route with JWT middleware
	r.GET("/users", middleware.ValidateToken, controllers.GetUser(db))
	// Protect /users/all route with API key middleware
	// This route is for admin access to get all users
	r.GET("/users/all", middleware.ValidateAPIKey, controllers.GetAllUsers(db))
	// Protect /products route with API key middleware
	// This route is for admin access to create and get products
	r.POST("/products", middleware.ValidateAPIKey, controllers.CreateProduct(db))
	r.DELETE("/products/:id", middleware.ValidateAPIKey, controllers.DeleteProduct(db))
	r.GET("/products", middleware.ValidateAPIKey, controllers.GetProducts(db))
}
