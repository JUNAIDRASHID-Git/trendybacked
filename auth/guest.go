package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"gorm.io/gorm"
)

// POST /auth/guest
func CreateGuestUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		guestID := "guest_" + generateRandomString(16)

		guest := models.GuestUser{
			ID:        guestID,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}

		if err := db.Create(&guest).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create guest"})
			return
		}

		// Issue JWT for guest
		token, err := issueGuestToken(guestID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Token generation failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"guest_id":   guestID,
			"token":      token,
			"expires_at": guest.ExpiresAt,
		})
	}
}

func generateRandomString(n int) string {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "rand_guest"
	}
	return hex.EncodeToString(bytes)
}

func issueGuestToken(id string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": id,
		"role":    "guest",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
