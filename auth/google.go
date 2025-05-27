package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"gorm.io/gorm"
)

// --- Google Token Info Struct ---
type GoogleTokenInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Aud           string `json:"aud"`
}

// --- Google Auth Handler ---
func GoogleAuthHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var body struct {
			IDToken string `json:"id_token"`
		}

		// Step 1: Get the ID token from request body
		if err := c.ShouldBindJSON(&body); err != nil || body.IDToken == "" {
			log.Println("⚠️  Missing or invalid ID token in request.")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid ID token"})
			return
		}

		log.Println("✅ Received ID token from client. Verifying...")

		// Step 2: Verify Google token
		userData, err := verifyGoogleToken(body.IDToken)
		if err != nil {
			log.Printf("❌ Google token verification failed: %v\n", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		log.Printf("✅ Google token verified for email: %s\n", userData.Email)

		// Step 3: Find or create the user
		user, err := findOrCreateUser(db, userData)
		if err != nil {
			log.Printf("❌ Database error: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		log.Printf("✅ User %s authenticated.\n", user.Email)

		// Step 4: Generate JWT token for the user
		token, err := GenerateJWT(user.ID, user.Email)
		if err != nil {
			log.Printf("❌ Failed to generate JWT: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		log.Printf("✅ JWT generated for user: %s\n", user.Email)

		// Step 5: Return the token to the client
		c.JSON(http.StatusOK, gin.H{"token": token})
	}
}

// --- Token Verification Function ---
func verifyGoogleToken(idToken string) (*GoogleTokenInfo, error) {
	resp, err := http.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + idToken)
	if err != nil {
		return nil, fmt.Errorf("failed to contact Google: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("invalid Google token: %s", body)
	}

	var tokenInfo GoogleTokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		return nil, fmt.Errorf("invalid token structure: %v", err)
	}

	expectedAud := os.Getenv("GOOGLE_CLIENT_ID")
	if expectedAud == "" {
		return nil, errors.New("missing GOOGLE_CLIENT_ID in environment variables")
	}

	if tokenInfo.Aud != expectedAud {
		return nil, errors.New("token audience mismatch")
	}

	return &tokenInfo, nil
}

// --- Create or Find User ---
func findOrCreateUser(db *gorm.DB, info *GoogleTokenInfo) (*models.User, error) {
	var user models.User
	result := db.First(&user, "id = ?", info.Sub)

	// If user is found, return the existing user
	if result.Error == nil {
		log.Printf("ℹ️  User %s found in database.\n", info.Email)
		return &user, nil
	}

	// If user not found, create a new user
	if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, result.Error
	}

	// Create a new user in the database
	user = models.User{
		ID:       info.Sub,
		Email:    info.Email,
		Name:     info.Name,
		Picture:  info.Picture,
		Provider: "google",
	}
	if err := db.Create(&user).Error; err != nil {
		return nil, err
	}

	log.Printf("✅ New user created: %s\n", user.Email)
	return &user, nil
}

// --- JWT Generation ---
func GenerateJWT(userID string, email string) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("missing JWT_SECRET in environment variables")
	}

	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
