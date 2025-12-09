package telrControllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	orderControllers "github.com/junaidrashid-git/ecommerce-api/controllers/order"
	"gorm.io/gorm"
)

// TelrPaymentResponse represents Telr response
type TelrPaymentResponse struct {
	Order struct {
		Ref string `json:"ref"`
		URL string `json:"url"`
	} `json:"order"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// getTelrConfig picks production endpoint, test mode if needed
func getTelrConfig() (storeID int, authKey, apiURL string, testMode int, err error) {
	storeID, _ = strconv.Atoi(os.Getenv("TELR_STORE_ID_PROD"))
	authKey = os.Getenv("TELR_AUTH_KEY_PROD")
	apiURL = os.Getenv("TELR_API_URL_PROD")
	testMode = 0

	mode := os.Getenv("TELR_MODE")
	if mode == "sandbox" || mode == "dev" {
		testMode = 1 // use test mode even on live endpoint
	}

	if storeID == 0 || authKey == "" || apiURL == "" {
		return 0, "", "", 0, fmt.Errorf("telr configuration missing")
	}
	return storeID, authKey, apiURL, testMode, nil
}

// CreateTelrPayment sends request to Telr and returns payment URL & order reference
func CreateTelrPayment(cartID, amount, currency, description, name, email, phone, addressLine1, addressLine2, city, region, country, postcode string) (string, string, error) {
	storeID, authKey, apiURL, testMode, err := getTelrConfig()
	if err != nil {
		return "", "", err
	}

	payload := map[string]interface{}{
		"method":  "create",
		"store":   storeID,
		"authkey": authKey,
		"order": map[string]interface{}{
			"cartid":      cartID,
			"test":        testMode,
			"amount":      amount,
			"currency":    currency,
			"description": description,
		},
		"customer": map[string]interface{}{
			"name":  name,
			"email": email,
			"phone": phone,
			"address": map[string]string{
				"line1":    addressLine1,
				"line2":    addressLine2,
				"city":     city,
				"region":   region,
				"country":  country,
				"postcode": postcode,
			},
		},
		"return": map[string]string{
			"authorised": os.Getenv("TELR_SUCCESS_URL"),
			"declined":   os.Getenv("TELR_FAILURE_URL"),
			"cancelled":  os.Getenv("TELR_CANCEL_URL"),
		},
	}

	jsonData, _ := json.Marshal(payload)
	fmt.Println("Telr Payload:", string(jsonData)) // debug log

	req, _ := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to reach Telr: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("telr API error (%d): %s", resp.StatusCode, string(body))
	}

	var telrResp TelrPaymentResponse
	if err := json.Unmarshal(body, &telrResp); err != nil {
		return "", "", fmt.Errorf("failed to parse Telr response: %v", err)
	}

	if telrResp.Error != nil {
		return "", "", fmt.Errorf("telr error: %s", telrResp.Error.Message)
	}

	if telrResp.Order.URL == "" {
		return "", "", fmt.Errorf("telr returned empty payment URL")
	}

	return telrResp.Order.URL, telrResp.Order.Ref, nil
}

// PaymentRequestHandler is the Gin handler
func PaymentRequestHandler(c *gin.Context) {
	var input struct {
		CartID      string `json:"cartid" binding:"required"`
		Amount      string `json:"amount" binding:"required"`
		Currency    string `json:"currency" binding:"required"`
		Description string `json:"description" binding:"required"`
		Name        string `json:"name" binding:"required"`
		Email       string `json:"email" binding:"required,email"`
		Phone       string `json:"phone" binding:"required"`
		// Optional: pass address from frontend
		AddressLine1 string `json:"address_line1"`
		AddressLine2 string `json:"address_line2"`
		City         string `json:"city"`
		Region       string `json:"region"`
		Country      string `json:"country"`
		Postcode     string `json:"postcode"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	fmt.Println("Incoming payment request:", input)

	paymentURL, orderRef, err := CreateTelrPayment(
		input.CartID,
		input.Amount,
		input.Currency,
		input.Description,
		input.Name,
		input.Email,
		input.Phone,
		input.AddressLine1,
		input.AddressLine2,
		input.City,
		input.Region,
		input.Country,
		input.Postcode,
	)

	if err != nil {
		c.JSON(502, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"payment_url": paymentURL,
		"order_ref":   orderRef,
	})
}

type TelrWebhookRequest struct {
	Order struct {
		Ref      string `json:"ref"`
		CartID   string `json:"cartid"`
		Amount   string `json:"amount"`
		Currency string `json:"currency"`
		Status   struct {
			Code int    `json:"code"` // 3 = Paid, 6 = Failed, etc.
			Text string `json:"text"`
		} `json:"status"`
		Customer struct {
			Name    string `json:"name"`
			Email   string `json:"email"`
			Phone   string `json:"phone"`
			Address struct {
				Line1    string `json:"line1"`
				Line2    string `json:"line2"`
				City     string `json:"city"`
				Region   string `json:"region"`
				Country  string `json:"country"`
				Postcode string `json:"postcode"`
			} `json:"address"`
		} `json:"customer"`
	} `json:"order"`
}

func TelrWebhookHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := c.Request.ParseForm(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse form"})
			return
		}

		fmt.Println("Received Telr webhook form:", c.Request.PostForm)

		cartID := c.PostForm("tran_cartid")
		tranStatus := c.PostForm("tran_status") // "A" = approved

		if cartID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing tran_cartid"})
			return
		}

		if tranStatus != "A" {
			c.JSON(http.StatusOK, gin.H{"message": "Payment not successful"})
			return
		}

		if err := orderControllers.PlaceOrder(db, cartID, "confirmed", "paid"); err != nil {
			fmt.Println("Failed to place order for cart:", cartID, "error:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to place order", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Order placed successfully"})
	}
}
