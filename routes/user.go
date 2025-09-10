package routes

import (
	"github.com/gin-gonic/gin"
	cartControllers "github.com/junaidrashid-git/ecommerce-api/controllers/cart"
	productControllers "github.com/junaidrashid-git/ecommerce-api/controllers/product"
	userControllers "github.com/junaidrashid-git/ecommerce-api/controllers/user"
	"github.com/junaidrashid-git/ecommerce-api/middleware"
	"gorm.io/gorm"
)

// SetupUserRoutes registers all “/user/*” endpoints. Requires JWT middleware.
func SetupUserRoutes(r *gin.Engine, db *gorm.DB) {
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

		// ──────────────── Browse Products ────────────────
		userGroup.GET("/products", productControllers.GetProducts(db))        // GET /user/products
		userGroup.GET("/products/:id", productControllers.GetProductByID(db)) // GET /user/products

		// ──────────────── Browse Categories + Products ────────────────
		userGroup.GET("/categories", userControllers.GetAllCategoriesWithProducts(db)) // GET /user/categories
	}
}
