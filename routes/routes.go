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

	// Google Admin login handler (wrapped for Gin)
	r.POST("/auth/google-admin", func(c *gin.Context) {
		auth.GoogleAdminLoginHandler(c.Writer, c.Request, db)
	})

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

		// all admin Fetching routes
		adminGroup.GET("/admins", func(c *gin.Context) {
			controllers.GetAllAdminsHandler(c.Writer, c.Request, db)
		})

		// all users fetching routes
		adminGroup.GET("/users", controllers.GetAllUsers(db))

		// products routes
		adminGroup.POST("/products", controllers.CreateProduct(db))
		adminGroup.PUT("/products/:id", controllers.UpdateProduct(db))
		adminGroup.GET("/products", controllers.GetProducts(db))
		adminGroup.DELETE("/products/:id", controllers.DeleteProduct(db))

		// excel import product routes
		adminGroup.POST("/products/import-excel", controllers.ImportProductsFromExcel(db))

		// Categores routes
		adminGroup.POST("/products/category", controllers.CreateCategory(db))
		adminGroup.PUT("/products/category", controllers.UpdateCategory(db))
		adminGroup.GET("/products/category", controllers.GetAllCategory(db))
		adminGroup.DELETE("/products/category/:id", controllers.DeleteCategory(db))

		// Admin management routes
		adminGroup.GET("/pending-admins", auth.ListPendingAdmins(db))
		adminGroup.POST("/approve-admin", auth.ApproveAdmin(db))
		adminGroup.POST("/reject-admin", auth.RejectAdmin(db))
	}
}
