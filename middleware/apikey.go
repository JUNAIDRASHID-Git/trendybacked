package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
)



func ValidateAPIKey(c *gin.Context) {
	apiKey := c.GetHeader("X-API-KEY")
	if apiKey != os.Getenv("COST_API_KEY") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or missing API key"})
		c.Abort()
		return
	}
	c.Next()
}
