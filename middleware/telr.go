package middleware

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// TelrWebhookAuth verifies Telr webhook signature, skips check in sandbox/dev mode
func TelrWebhookAuth() gin.HandlerFunc {
	secretKey := os.Getenv("TELR_WEBHOOK_SECRET")
	if secretKey == "" {
		panic("TELR_WEBHOOK_SECRET is not set")
	}

	mode := strings.ToLower(os.Getenv("TELR_MODE")) // "sandbox" or "dev" or "production"

	return func(c *gin.Context) {
		// Allow sandbox/dev to bypass verification for local testing
		if mode == "sandbox" || mode == "dev" {
			fmt.Println("Sandbox/dev mode: skipping Telr webhook signature verification")
			c.Next()
			return
		}

		// Parse form so we can access tran_check and other fields
		if err := c.Request.ParseForm(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse form for signature verification"})
			c.Abort()
			return
		}

		// Telr sends the SHA1 check in tran_check (or card_check / bill_check depending on payload)
		providedCheck := c.PostForm("tran_check")
		if providedCheck == "" {
			// If tran_check isn't present, fail verification (could log & inspect the payload)
			c.JSON(http.StatusForbidden, gin.H{"error": "missing tran_check signature"})
			c.Abort()
			return
		}

		// The list of fields Telr uses for the transaction check (from Telr docs).
		// If you have custom configuration, ensure this list matches what Telr is sending.
		fieldList := []string{
			"tran_store", "tran_type", "tran_class", "tran_test", "tran_ref",
			"tran_prevref", "tran_firstref", "tran_order", "tran_currency",
			"tran_amount", "tran_cartid", "tran_desc", "tran_status",
			"tran_authcode", "tran_authmessage",
		}

		// Build signature string: secretkey + ":" + valueForEachField (empty if missing)
		var parts []string
		parts = append(parts, secretKey)
		for _, f := range fieldList {
			v := strings.TrimSpace(c.PostForm(f))
			parts = append(parts, v) // empty string if missing (Telr expects that)
		}
		signatureString := strings.Join(parts, ":")

		// compute SHA1 hex
		h := sha1.New()
		h.Write([]byte(signatureString))
		calculated := hex.EncodeToString(h.Sum(nil))

		fmt.Println("Telr provided tran_check:", providedCheck)
		fmt.Println("Telr calculated SHA1:", calculated)

		if !strings.EqualFold(calculated, providedCheck) {
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid webhook signature"})
			c.Abort()
			return
		}

		// OK
		c.Next()
	}
}
