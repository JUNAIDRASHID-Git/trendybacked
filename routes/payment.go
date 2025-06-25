package routes

import (
	"github.com/gin-gonic/gin"
	paymentControllers "github.com/junaidrashid-git/ecommerce-api/controllers/tap_payment"
)

func SetupPaymentRoutes(r *gin.Engine) {
	payments := r.Group("/api/payment")
	{
		// Initialize a new payment
		payments.POST("/init", paymentControllers.InitTapPaymentHandler())
		payments.POST("/webhook", paymentControllers.TapWebhookHandler())
		payments.POST("/status", paymentControllers.CheckPaymentStatusHandler())
	}
}
