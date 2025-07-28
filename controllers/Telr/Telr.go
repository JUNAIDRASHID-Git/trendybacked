package paymentTelrControllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
)

type TelrPaymentRequest struct {
	Amount        float64 `json:"amount"`
	UserID        string  `json:"user_id"`
	CustomerEmail string  `json:"customer_email"`
}

// InitTelrPaymentHandler initializes Telr payment request
func InitTelrPaymentHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req TelrPaymentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload"})
			return
		}

		storeID := os.Getenv("TELR_STORE_ID")
		authKey := os.Getenv("TELR_AUTH_KEY")
		if storeID == "" || authKey == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Telr credentials not set"})
			return
		}

		params := map[string]string{
			"ivp_method":   "create",
			"ivp_store":    storeID,
			"ivp_authkey":  authKey,
			"ivp_currency": "SAR",
			"ivp_amount":   fmt.Sprintf("%.2f", req.Amount),
			"ivp_test":     "0",
			"ivp_cart":     req.UserID,
			"ivp_desc":     "TrendyChef Payment",
			"return_auth":  "https://yourdomain.com/payment/return_success",
			"return_decl":  "https://yourdomain.com/payment/return_failed",
			"return_can":   "https://yourdomain.com/payment/return_cancelled",
			"bill_email":   req.CustomerEmail,
		}

		form := url.Values{}
		for k, v := range params {
			form.Set(k, v)
		}

		resp, err := http.PostForm("https://secure.telr.com/gateway/order.json", form)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Telr request failed"})
			return
		}
		defer resp.Body.Close()

		var result struct {
			Method string `json:"method"`
			Order  struct {
				Ref string `json:"ref"`
				URL string `json:"url"`
			} `json:"order"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid Telr response"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"order_ref":    result.Order.Ref,
			"redirect_url": result.Order.URL,
		})
	}
}

// CheckPaymentStatusHandler checks the status of a Telr payment
func CheckPaymentStatusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			OrderRef string `json:"order_ref" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "order_ref required"})
			return
		}

		storeID := os.Getenv("TELR_STORE_ID")
		authKey := os.Getenv("TELR_AUTH_KEY")
		params := url.Values{
			"ivp_method":  {"check"},
			"ivp_store":   {storeID},
			"ivp_authkey": {authKey},
			"order_ref":   {req.OrderRef},
		}

		resp, err := http.PostForm("https://secure.telr.com/gateway/order.json", params)
		if err != nil || resp.StatusCode != http.StatusOK {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Telr status request failed"})
			return
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid Telr status response"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"result": result})
	}
}
