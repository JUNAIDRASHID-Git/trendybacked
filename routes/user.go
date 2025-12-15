package routes

import (
	"github.com/gin-gonic/gin"
	cartControllers "github.com/junaidrashid-git/ecommerce-api/controllers/cart"
	productControllers "github.com/junaidrashid-git/ecommerce-api/controllers/product"
	userControllers "github.com/junaidrashid-git/ecommerce-api/controllers/user"
	"github.com/junaidrashid-git/ecommerce-api/middleware"
	"gorm.io/gorm"
)

// SetupUserRoutes registers all “/user/*” endpoints.
// User & Cart require JWT; Products & Categories are PUBLIC.
func SetupUserRoutes(r *gin.Engine, db *gorm.DB) {

	guestGroup := r.Group("/guest")
	{
		guestGroup.GET("/cart", cartControllers.GetGuestCart(db)) // GET /guest/cart?guest_id=xxx
		guestGroup.POST("/cart", cartControllers.UpdateGuestCartItem(db))
		guestGroup.DELETE("/cart/:product_id", cartControllers.DeleteGuestCartItem(db))
		guestGroup.DELETE("/cart", cartControllers.ClearGuestCart(db)) // DELETE /guest/cart
	}

	// ──────────────── PUBLIC ROUTES ────────────────
	publicGroup := r.Group("/public")
	{
		// Publicly accessible product routes
		publicGroup.GET("/products", productControllers.GetProducts(db))                // GET /public/products
		publicGroup.GET("/products/:id", productControllers.GetProductByID(db))         // GET /public/products/:id
		publicGroup.GET("/og/products/:id", productControllers.GetProductOGHandler(db)) // GET /og/public/products/:id
		// Publicly accessible category routes
		publicGroup.GET("/categories", productControllers.GetAllCategoriesWithProducts(db))
		publicGroup.GET("/categories/:id", productControllers.GetCategoryByID(db))
	}

	// ──────────────── AUTHENTICATED USER ROUTES ────────────────
	userGroup := r.Group("/user")
	userGroup.Use(middleware.ValidateToken)
	{
		// ──────────────── User Profile ────────────────
		userGroup.GET("/", userControllers.GetUser(db))    // GET /user/
		userGroup.PUT("/", userControllers.UpdateUser(db)) // PUT /user/

		// ──────────────── Shopping Cart ────────────────
		cartGroup := userGroup.Group("/cart")
		{
			cartGroup.GET("/", cartControllers.GetUserCart(db))                  // GET /user/cart
			cartGroup.POST("/", cartControllers.UpdateCartItem(db))              // POST /user/cart
			cartGroup.DELETE("/:product_id", cartControllers.DeleteCartItem(db)) // DELETE /user/cart/:product_id
			cartGroup.DELETE("/", cartControllers.ClearUserCart(db))             // DELETE /user/cart
		}
	}
}
