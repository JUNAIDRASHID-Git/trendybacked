package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRoutes is the single entry‐point that wires up Auth, User, and Admin route groups.
func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	// 1️⃣ Public Auth routes (no middleware)
	SetupAuthRoutes(r, db)

	// 2️⃣ User routes (JWT‐protected)
	SetupUserRoutes(r, db)

	// 3️⃣ Admin routes (API‐Key‐protected)
	SetupAdminRoutes(r, db)

	// order routes
	SetupOrderRoutes(r, db)

	// telr payment routes

	SetupTelrRoutes(r, db)
}
