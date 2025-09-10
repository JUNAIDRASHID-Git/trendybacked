package routes

import (
	"github.com/gin-gonic/gin"
	telrControllers "github.com/junaidrashid-git/ecommerce-api/controllers/telr"
	"github.com/junaidrashid-git/ecommerce-api/middleware"
	"gorm.io/gorm"
)

func SetupTelrRoutes(r *gin.Engine, db *gorm.DB) {
	payment := r.Group("/payment")
	{
		// Payment creation endpoint
		payment.POST("/place", telrControllers.PaymentRequestHandler)

		// Webhook endpoint: middleware handles sandbox/prod verification
		payment.POST("/webhook",
			middleware.TelrWebhookAuth(),
			telrControllers.TelrWebhookHandler(db),
		)
	}
}
