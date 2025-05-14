package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing or invalid ID token"})
			return
		}

		// Step 2: Verify Google token
		userData, err := verifyGoogleToken(body.IDToken)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		// Step 3: Find or create the user
		user, err := findOrCreateUser(db, userData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}

		// Step 4: Generate JWT token for the user
		token, err := GenerateJWT(user.ID, user.Email)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		// Step 5: Return the token to the client
		c.JSON(http.StatusOK, gin.H{"token": token})
	}
}

// --- Token Verification Function ---
func verifyGoogleToken(idToken string) (*GoogleTokenInfo, error) {
	// Make a GET request to verify the token using Google's OAuth2 API
	resp, err := http.Get("https://oauth2.googleapis.com/tokeninfo?id_token=" + idToken)
	if err != nil {
		return nil, fmt.Errorf("failed to contact Google: %v", err)
	}
	defer resp.Body.Close()

	// Check if the response is successful
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("invalid Google token: %s", body)
	}

	// Parse the response JSON into the GoogleTokenInfo struct
	var tokenInfo GoogleTokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		return nil, fmt.Errorf("invalid token structure: %v", err)
	}

	// Ensure the token's audience matches the expected Google Client ID
	expectedAud := os.Getenv("GOOGLE_CLIENT_ID")
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

	return &user, nil
}

// --- JWT Generation ---
func GenerateJWT(userID string, email string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	return token.SignedString([]byte(secret))
}
