package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	firebase "firebase.google.com/go"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"google.golang.org/api/option"
	"gorm.io/gorm"
)

// InitFirebase initializes Firebase app and auth client
func InitFirebase() {
	ctx := context.Background()

	credsJSON := os.Getenv("FIREBASE_CREDENTIALS_JSON")
	if credsJSON == "" {
		log.Fatal("FIREBASE_CREDENTIALS_JSON must be set")
	}

	projectID = os.Getenv("FIREBASE_PROJECT_ID")
	if projectID == "" {
		log.Fatal("FIREBASE_PROJECT_ID must be set")
	}

	opt := option.WithCredentialsJSON([]byte(credsJSON))
	config := &firebase.Config{ProjectID: projectID}

	var err error
	firebaseApp, err = firebase.NewApp(ctx, config, opt)
	if err != nil {
		log.Fatalf("Error initializing Firebase app: %v", err)
	}

	firebaseAuth, err = firebaseApp.Auth(ctx)
	if err != nil {
		log.Fatalf("Error getting Firebase Auth client: %v", err)
	}
}

// GoogleUserLoginHandler handles login/signup via Google OAuth ID token
func GoogleUserLoginHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	var req struct {
		IDToken string `json:"idToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.IDToken == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	token, err := firebaseAuth.VerifyIDTokenAndCheckRevoked(ctx, req.IDToken)
	if err != nil {
		log.Printf("ID token verification failed: %v", err)
		http.Error(w, "Invalid or revoked ID token", http.StatusUnauthorized)
		return
	}

	if token.Audience != projectID {
		log.Printf("Token audience mismatch: got %q", token.Audience)
		http.Error(w, "Invalid token audience", http.StatusUnauthorized)
		return
	}

	email, ok := token.Claims["email"].(string)
	if !ok || email == "" {
		http.Error(w, "Email not found in token", http.StatusUnauthorized)
		return
	}
	name, _ := token.Claims["name"].(string)
	picture, _ := token.Claims["picture"].(string)

	firebaseUserID := token.UID

	var user models.User
	err = db.First(&user, "id = ?", firebaseUserID).Error
	if err == gorm.ErrRecordNotFound {
		user = models.User{
			ID:       firebaseUserID,
			Email:    email,
			Name:     name,
			Picture:  picture,
			Provider: "google",
			Cart:     models.Cart{UserID: firebaseUserID},
		}

		// ✅ FIX: removed .Omit("ID")
		if err := db.Create(&user).Error; err != nil {
			log.Printf("❌ Failed to register user: %v", err)
			http.Error(w, "Failed to register user", http.StatusInternalServerError)
			return
		}
		log.Printf("✅ New user registered: %s", email)
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	} else {
		// Update user info if user already exists
		updates := models.User{
			Name:    name,
			Picture: picture,
		}
		if err := db.Model(&user).Updates(updates).Error; err != nil {
			http.Error(w, "Failed to update user info", http.StatusInternalServerError)
			return
		}
		log.Printf("✅ Existing user updated: %s", email)
	}

	// Issue a JWT token and respond
	issueTokenAndRespond(w, email, "user", firebaseUserID, name, picture)
}

// func GoogleUserLogoutHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
// 	var req struct {
// 		UserID string `json:"userId"`
// 	}

// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserID == "" {
// 		http.Error(w, "Invalid request", http.StatusBadRequest)
// 		return
// 	}

// 	ctx := context.Background()

// 	err := firebaseAuth.RevokeRefreshTokens(ctx, req.UserID)
// 	if err != nil {
// 		log.Printf("❌ Failed to revoke tokens: %v", err)
// 		http.Error(w, "Failed to revoke user tokens", http.StatusInternalServerError)
// 		return
// 	}

// 	log.Printf("✅ Refresh tokens revoked for user: %s", req.UserID)
// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(map[string]string{"message": "User logged out"})
// }
