package paymentControllers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// TapPaymentRequest defines the expected request body for initiating payment
type TapPaymentRequest struct {
	Amount        float64 `json:"amount"`
	UserID        string  `json:"user_id"`
	CustomerEmail string  `json:"customer_email"`
}

// InitTapPaymentHandler handles payment initialization with Tap
func InitTapPaymentHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		secretKey := os.Getenv("TAP_SECRET_KEY")
		if secretKey == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Tap secret key not set in environment"})
			return
		}

		var req TapPaymentRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
			return
		}

		payload := map[string]interface{}{
			"amount":               req.Amount,
			"currency":             "SAR",
			"threeDSecure":         true,
			"save_card":            false,
			"description":          "TrendyChef Payment",
			"statement_descriptor": "TrendyChef",
			"customer": map[string]interface{}{
				"first_name": "Tap",
				"last_name":  "User",
				"email":      req.CustomerEmail,
			},
			"source": map[string]string{
				"id": "src_all",
			},
			"post": map[string]interface{}{
				"url": "http://localhost:8080/api/payment/webhook", // Change this to your live endpoint
			},
			"redirect": map[string]interface{}{
				"url": "http://localhost:443/#/payment-redirect",
			},
			"metadata": map[string]interface{}{
				"user_id": req.UserID,
			},
		}

		body, err := json.Marshal(payload)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode request body"})
			return
		}

		reqTap, err := http.NewRequest("POST", "https://api.tap.company/v2/charges", bytes.NewBuffer(body))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request to Tap"})
			return
		}
		reqTap.Header.Set("Authorization", "Bearer "+secretKey)
		reqTap.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(reqTap)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to Tap"})
			return
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse Tap response"})
			return
		}

		// Return redirect URL to client
		if transaction, ok := result["transaction"].(map[string]interface{}); ok {
			if url, exists := transaction["url"].(string); exists {
				c.JSON(http.StatusOK, gin.H{"redirect_url": url})
				return
			}
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Transaction URL not found",
			"details": result,
		})
	}
}

// TapWebhookHandler verifies and processes webhook events from Tap
func TapWebhookHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
			return
		}

		tapSignature := c.GetHeader("X-Tap-Signature")
		if tapSignature == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing X-Tap-Signature header"})
			return
		}

		webhookSecret := os.Getenv("TAP_WEBHOOK_SECRET")
		if webhookSecret == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Webhook secret not set in environment"})
			return
		}

		// Verify the HMAC signature
		mac := hmac.New(sha256.New, []byte(webhookSecret))
		mac.Write(body)
		expected := hex.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(expected), []byte(tapSignature)) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid signature"})
			return
		}

		var event map[string]interface{}
		if err := json.Unmarshal(body, &event); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON in webhook body"})
			return
		}

		eventType, ok := event["type"].(string)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing event type"})
			return
		}

		data, ok := event["data"].(map[string]interface{})
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing data object"})
			return
		}

		// Handle event type
		switch eventType {
		case "charge.succeeded":
			fmt.Println("✅ Payment succeeded. Charge ID:", data["id"])
			// TODO: Create order in DB, mark payment complete
		case "charge.failed":
			fmt.Println("❌ Payment failed. Charge ID:", data["id"])
			// TODO: Log failure, notify user if needed
		default:
			fmt.Println("ℹ️ Unhandled event:", eventType)
		}

		c.Status(http.StatusOK)
	}
}

// CheckPaymentStatusHandler checks the status of a charge by ID
func CheckPaymentStatusHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			ChargeID string `json:"charge_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Charge ID is required"})
			return
		}

		secretKey := os.Getenv("TAP_SECRET_KEY")
		if secretKey == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Secret key not set in environment"})
			return
		}

		url := fmt.Sprintf("https://api.tap.company/v2/charges/%s", req.ChargeID)
		reqTap, err := http.NewRequest("GET", url, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}
		reqTap.Header.Set("Authorization", "Bearer "+secretKey)

		client := &http.Client{}
		resp, err := client.Do(reqTap)
		if err != nil || resp.StatusCode != http.StatusOK {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve payment status"})
			return
		}
		defer resp.Body.Close()

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
			return
		}

		status, _ := result["status"].(string)
		c.JSON(http.StatusOK, gin.H{
			"status": status,
			"data":   result,
		})
	}
}
