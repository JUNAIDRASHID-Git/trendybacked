package routes

import (
	"github.com/gin-gonic/gin"
	paymentControllers "github.com/junaidrashid-git/ecommerce-api/controllers/telr_payment"
)

func SetupPaymentRoutes(r *gin.Engine) {
	payments := r.Group("/api/payment")
	{
		// Initialize a new payment
		payments.POST("/init", paymentControllers.InitTelrPaymentHandler())
		payments.POST("/status", paymentControllers.CheckPaymentStatusHandler())
	}
}
