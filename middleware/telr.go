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

	mode := strings.ToLower(os.Getenv("TELR_MODE"))

	return func(c *gin.Context) {
		if mode == "sandbox" || mode == "dev" {
			fmt.Println("Sandbox/dev mode: skipping Telr webhook signature verification")
			c.Next()
			return
		}

		if err := c.Request.ParseForm(); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse form for signature verification"})
			c.Abort()
			return
		}

		providedCheck := c.PostForm("tran_check")
		if providedCheck == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "missing tran_check signature"})
			c.Abort()
			return
		}

		fieldList := []string{
			"tran_store", "tran_type", "tran_class", "tran_test", "tran_ref",
			"tran_prevref", "tran_firstref", "tran_order", "tran_currency",
			"tran_amount", "tran_cartid", "tran_desc", "tran_status",
			"tran_authcode", "tran_authmessage",
		}

		parts := []string{secretKey}
		for _, f := range fieldList {
			v := strings.TrimSpace(c.PostForm(f))
			parts = append(parts, v)
		}

		signatureString := strings.Join(parts, ":")
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

		c.Next()
	}
}
