package telrControllers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
		return 0, "", "", 0, fmt.Errorf("Telr configuration missing")
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

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal Telr payload: %v", err)
	}
	fmt.Println("Telr Payload:", string(jsonData)) // debug log

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", "", fmt.Errorf("failed to build Telr request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to reach Telr: %v", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("Telr API error (%d): %s", resp.StatusCode, string(body))
	}

	var telrResp TelrPaymentResponse
	if err := json.Unmarshal(body, &telrResp); err != nil {
		return "", "", fmt.Errorf("failed to parse Telr response: %v", err)
	}

	if telrResp.Error != nil {
		return "", "", fmt.Errorf("Telr error: %s", telrResp.Error.Message)
	}

	if telrResp.Order.URL == "" {
		return "", "", fmt.Errorf("Telr returned empty payment URL")
	}

	return telrResp.Order.URL, telrResp.Order.Ref, nil
}

// PaymentRequestHandler is the Gin handler
func PaymentRequestHandler(c *gin.Context) {
	// Enforce Content-Type: application/json (accepts charset variants)
	ct := c.GetHeader("Content-Type")
	if ct == "" || !strings.HasPrefix(strings.ToLower(ct), "application/json") {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"error":   "invalid content-type",
			"details": "Content-Type must be application/json",
		})
		return
	}

	// Read raw body for logging and fallback parsing, and restore it for binding
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}
	// restore for ShouldBindJSON
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Log raw body (dev only; remove or redact in production)
	fmt.Println("Raw request body:", string(bodyBytes))

	var input struct {
		CartID       string `json:"cartid" binding:"required"`
		Amount       string `json:"amount" binding:"required"`
		Currency     string `json:"currency" binding:"required"`
		Description  string `json:"description" binding:"required"`
		Name         string `json:"name" binding:"required"`
		Email        string `json:"email" binding:"required,email"`
		Phone        string `json:"phone" binding:"required"`
		AddressLine1 string `json:"address_line1"`
		AddressLine2 string `json:"address_line2"`
		City         string `json:"city"`
		Region       string `json:"region"`
		Country      string `json:"country"`
		Postcode     string `json:"postcode"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		// Log bind error
		fmt.Println("Bind error:", err.Error())

		// Fallback: try to unmarshal into a map and pick alternate keys (cart_id, cartId)
		var m map[string]interface{}
		if err2 := json.Unmarshal(bodyBytes, &m); err2 == nil {
			// Helper to read string values
			getStr := func(keys ...string) string {
				for _, k := range keys {
					if v, ok := m[k]; ok {
						switch t := v.(type) {
						case string:
							if t != "" {
								return t
							}
						case float64:
							// numeric values can be formatted
							return fmt.Sprintf("%v", t)
						}
					}
				}
				return ""
			}

			// fill possible alternate keys
			if input.CartID == "" {
				input.CartID = getStr("cartid", "cart_id", "cartId")
			}
			if input.Amount == "" {
				input.Amount = getStr("amount")
			}
			if input.Currency == "" {
				input.Currency = getStr("currency")
			}
			if input.Description == "" {
				input.Description = getStr("description")
			}
			if input.Name == "" {
				input.Name = getStr("name")
			}
			if input.Email == "" {
				input.Email = getStr("email")
			}
			if input.Phone == "" {
				input.Phone = getStr("phone")
			}
			if input.AddressLine1 == "" {
				input.AddressLine1 = getStr("address_line1", "addressLine1", "line1")
			}
			if input.AddressLine2 == "" {
				input.AddressLine2 = getStr("address_line2", "addressLine2", "line2")
			}
			if input.City == "" {
				input.City = getStr("city")
			}
			if input.Region == "" {
				input.Region = getStr("region")
			}
			if input.Country == "" {
				input.Country = getStr("country")
			}
			if input.Postcode == "" {
				input.Postcode = getStr("postcode", "postal_code")
			}

			// Basic required fields check after fallback
			if input.CartID == "" || input.Amount == "" || input.Currency == "" || input.Description == "" || input.Name == "" || input.Email == "" || input.Phone == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "invalid request",
					"details": err.Error(),
				})
				return
			}

			// Basic email validation
			if _, emailErr := mail.ParseAddress(input.Email); emailErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "invalid email",
					"details": emailErr.Error(),
				})
				return
			}
		} else {
			// cannot parse JSON at all
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "invalid JSON",
				"details": err.Error(),
			})
			return
		}
	}

	// At this point we have a validated input (either from binding or fallback)
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
		fmt.Println("CreateTelrPayment error:", err.Error())
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"payment_url": paymentURL,
		"order_ref":   orderRef,
	})
}

func TelrWebhookHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Telr sends form-urlencoded payload
		if err := c.Request.ParseForm(); err != nil {
			fmt.Println("[Webhook] ❌ Failed to parse form:", err)
			c.JSON(400, gin.H{"error": "failed to parse form"})
			return
		}

		tranStatus := c.PostForm("tran_status")
		cartID := c.PostForm("tran_cartid")
		orderRef := c.PostForm("tran_order")
		amountStr := c.PostForm("tran_amount")

		fmt.Printf("[Webhook] Incoming Telr webhook: status=%s cartID=%s orderRef=%s amount=%s\n",
			tranStatus, cartID, orderRef, amountStr)

		if tranStatus != "A" {
			fmt.Println("[Webhook] ⚠️ Payment not authorized for cart:", cartID)
			c.JSON(200, gin.H{"message": "payment not authorized"})
			return
		}

		var paidAmount float64
		fmt.Sscanf(amountStr, "%f", &paidAmount)

		// Start transaction
		err := db.Transaction(func(tx *gorm.DB) error {
			// Fetch cart with items
			var cart models.Cart
			if err := tx.Preload("Items").Where("cart_id = ?", cartID).First(&cart).Error; err != nil {
				return err
			}

			if len(cart.Items) == 0 {
				return errors.New("cart is empty")
			}

			var orderItems []models.OrderItem
			var total, totalWeight float64

			for _, item := range cart.Items {
				// Lock product row for update
				var product models.Product
				if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
					First(&product, "id = ?", item.ProductID).Error; err != nil {
					return err
				}

				// Check stock
				if product.Stock < item.Quantity {
					return fmt.Errorf("insufficient stock for product: %s", item.ProductEName)
				}

				// Decrement stock
				product.Stock -= item.Quantity
				if err := tx.Save(&product).Error; err != nil {
					return err
				}

				total += item.ProductSalePrice * float64(item.Quantity)
				totalWeight += item.Weight * float64(item.Quantity)

				orderItems = append(orderItems, models.OrderItem{
					ProductID:           item.ProductID,
					ProductEName:        item.ProductEName,
					ProductArName:       item.ProductArName,
					ProductImage:        item.ProductImage,
					ProductSalePrice:    item.ProductSalePrice,
					ProductRegularPrice: item.ProductRegularPrice,
					Weight:              item.Weight,
					Quantity:            item.Quantity,
				})
			}

			// Shipping cost calculation
			shippingCost := 0.0
			if totalWeight > 0 {
				shippingCost = float64(int(math.Ceil((totalWeight-1)/30.0))) * 30.0
			}

			order := models.Order{
				UserID:        cart.UserID,
				Items:         orderItems,
				TotalAmount:   total + shippingCost,
				ShippingCost:  shippingCost,
				PaymentStatus: models.PaymentStatusPaid,
				PaymentMethod: "card",
				OrderRef:      orderRef,
				Status:        models.OrderStatusConfirmed,
				CreatedAt:     time.Now(),
			}

			if err := tx.Create(&order).Error; err != nil {
				return err
			}

			// Clear cart
			if err := tx.Where("cart_id = ?", cart.CartID).Delete(&models.CartItem{}).Error; err != nil {
				return err
			}

			fmt.Printf("[Webhook] ✅ Order created successfully. OrderRef=%s, UserID=%s, Amount=%.2f\n",
				orderRef, cart.UserID, order.TotalAmount)

			return nil
		})

		if err != nil {
			fmt.Printf("[Webhook] ❌ Failed to create order. CartID=%s, Error=%v\n", cartID, err)
			c.JSON(500, gin.H{"error": "failed to create order", "details": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "order created successfully"})
	}
}
