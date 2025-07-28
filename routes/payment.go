package routes

import (
	"github.com/gin-gonic/gin"
	paymentTelrControllers "github.com/junaidrashid-git/ecommerce-api/controllers/telr"
)

func SetupPaymentRoutes(r *gin.Engine) {
	payments := r.Group("/api/payment")
	{
		// Initialize a new payment
		payments.POST("/init", paymentTelrControllers.InitTelrPaymentHandler())
		payments.POST("/status", paymentTelrControllers.CheckPaymentStatusHandler())
	}
}
