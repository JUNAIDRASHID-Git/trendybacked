package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/auth"
	"gorm.io/gorm"
)

// SetupAuthRoutes registers all “/auth/*” endpoints.
func SetupAuthRoutes(r *gin.Engine, db *gorm.DB) {
	authGroup := r.Group("/auth")
	{
		// Regular user Google login
		authGroup.POST("/google-user", func(c *gin.Context) {
			auth.GoogleUserLoginHandler(c.Writer, c.Request, db)
		})

		// Google Admin login (wrapped as a Gin handler)
		authGroup.POST("/google-admin", func(c *gin.Context) {
			auth.GoogleAdminLoginHandler(c.Writer, c.Request, db)
		})

		authGroup.POST("/guest", auth.CreateGuestUser(db))
	}
}
