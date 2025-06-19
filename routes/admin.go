package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/controllers/admin"
	"github.com/junaidrashid-git/ecommerce-api/controllers/product"
	"github.com/junaidrashid-git/ecommerce-api/controllers/user"
	"github.com/junaidrashid-git/ecommerce-api/middleware"
	"gorm.io/gorm"
)

// SetupAdminRoutes registers all “/admin/*” endpoints. Requires API‐Key middleware.
func SetupAdminRoutes(r *gin.Engine, db *gorm.DB) {
	adminGroup := r.Group("/admin")
	adminGroup.Use(middleware.ValidateAPIKey)
	{
		// ─────────── Admin & User Management ───────────
		adminGroup.GET("/admins", adminController.GetAllAdmins(db))
		adminGroup.GET("/users", userControllers.GetAllUsers(db))

		// ─────────── Product Management ───────────
		productAdmin := adminGroup.Group("/products")
		{
			productAdmin.POST("", productcontroller.CreateProduct(db))
			productAdmin.PUT("/:id", productcontroller.UpdateProduct(db))
			productAdmin.GET("", productcontroller.GetProducts(db))
			productAdmin.DELETE("/:id", productcontroller.DeleteProduct(db))
			productAdmin.POST("/import-excel", productcontroller.ImportProductsFromExcel(db))
		}

		// ─────────── Category Management ───────────
		categoryAdmin := adminGroup.Group("/categories")
		{
			categoryAdmin.POST("", productcontroller.CreateCategory(db))
			categoryAdmin.PUT("/:id", productcontroller.UpdateCategory(db))
			categoryAdmin.GET("", productcontroller.GetAllCategories(db))
			categoryAdmin.DELETE("/:id", productcontroller.DeleteCategory(db))
		}

		// ─────────── Admin Approval Workflow ───────────
		adminMgmt := adminGroup.Group("/admin-management")
		{
			adminMgmt.GET("/pending", adminController.ListPendingAdmins(db))
			adminMgmt.POST("/approve", adminController.ApproveAdmin(db))
			adminMgmt.POST("/reject", adminController.RejectAdmin(db))
		}

		bannerMgmt := adminGroup.Group("/banner")
		{
			bannerMgmt.POST("/upload", adminController.UploadBanner(db))
			bannerMgmt.GET("/", adminController.GetBanners(db))
			bannerMgmt.DELETE("/", adminController.DeleteBanner(db))
		}
	}
}
