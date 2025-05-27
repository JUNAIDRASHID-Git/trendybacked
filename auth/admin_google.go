package auth

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"gorm.io/gorm"
)

var (
	firebaseApp  *firebase.App
	firebaseAuth *auth.Client
	projectID    string
)

func init() {
	// Load .env file if present
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️  No .env file found, proceeding with environment variables")
	}

	ctx := context.Background()

	// If FIREBASE_CREDENTIALS_PATH is not set, fall back to GOOGLE_APPLICATION_CREDENTIALS
	credsPath := os.Getenv("FIREBASE_CREDENTIALS_JSON")
	if credsPath == "" {
		credsPath = os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	}
	if credsPath == "" {
		log.Fatal("❌ FIREBASE_CREDENTIALS_PATH or GOOGLE_APPLICATION_CREDENTIALS must be set to your service account JSON")
	}

	projectID = os.Getenv("FIREBASE_PROJECT_ID")
	if projectID == "" {
		log.Fatal("❌ FIREBASE_PROJECT_ID must be set")
	}

	// Initialize Firebase
	opt := option.WithCredentialsFile(credsPath)
	config := &firebase.Config{ProjectID: projectID}

	var err error
	firebaseApp, err = firebase.NewApp(ctx, config, opt)
	if err != nil {
		log.Fatalf("❌ Error initializing Firebase app: %v", err)
	}

	firebaseAuth, err = firebaseApp.Auth(ctx)
	if err != nil {
		log.Fatalf("❌ Error getting Firebase Auth client: %v", err)
	}
}

// GoogleAdminLoginHandler handles admin login via Google OAuth2.
func GoogleAdminLoginHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	var req struct {
		IDToken string `json:"idToken"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.IDToken == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Verify the token AND check for revocation
	token, err := firebaseAuth.VerifyIDTokenAndCheckRevoked(ctx, req.IDToken)
	if err != nil {
		log.Printf("❌ ID token verification failed: %v", err)
		http.Error(w, "Invalid or revoked ID token", http.StatusUnauthorized)
		return
	}

	// (Optional) Double-check the audience and issuer
	if token.Audience != projectID {
		log.Printf("❌ Token audience mismatch: got %q", token.Audience)
		http.Error(w, "Invalid token audience", http.StatusUnauthorized)
		return
	}

	// Extract standard claims
	email, ok := token.Claims["email"].(string)
	if !ok || email == "" {
		http.Error(w, "Email not found in token", http.StatusUnauthorized)
		return
	}
	name, _ := token.Claims["name"].(string)
	picture, _ := token.Claims["picture"].(string)

	superAdminEmail := os.Getenv("SUPER_ADMIN_EMAIL")

	// Super admin shortcut
	if email == superAdminEmail {
		issueTokenAndRespond(w, email, "superadmin", name, picture)
		return
	}

	// Regular admin flow
	var admin models.Admin
	err = db.Where("email = ?", email).First(&admin).Error
	if err == gorm.ErrRecordNotFound {
		// Create pending admin
		admin = models.Admin{
			Email:    email,
			Name:     name,
			Picture:  picture,
			Approved: false,
		}
		if err := db.Create(&admin).Error; err != nil {
			http.Error(w, "Failed to register admin", http.StatusInternalServerError)
			return
		}
		log.Printf("📝 New admin registered: %s (pending approval)", email)
		http.Error(w, "Pending approval by super admin", http.StatusForbidden)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Update profile if changed
	if err := db.Model(&admin).Updates(models.Admin{Name: name, Picture: picture}).Error; err != nil {
		http.Error(w, "Failed to update admin info", http.StatusInternalServerError)
		return
	}

	// Reload to get the latest Approved flag
	if err := db.First(&admin, admin.ID).Error; err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if !admin.Approved {
		http.Error(w, "Pending approval by super admin", http.StatusForbidden)
		return
	}

	// Approved admin
	issueTokenAndRespond(w, email, "admin", name, picture)
}

// issueTokenAndRespond DRYs up issuing the JWT, setting the cookie (if you choose), and JSON.
func issueTokenAndRespond(w http.ResponseWriter, email, role, name, picture string) {
	jwtStr := generateJWT(email, role)

	// Optionally set as HttpOnly cookie here
	// http.SetCookie(w, &http.Cookie{ /* ... */ })

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"token":   jwtStr,
		"role":    role,
		"email":   email,
		"name":    name,
		"picture": picture,
	})
}

func generateJWT(email, role string) string {
	claims := jwt.MapClaims{
		"email": email,
		"role":  role,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	hmac := []byte(os.Getenv("JWT_SECRET"))
	signed, err := t.SignedString(hmac)
	if err != nil {
		log.Printf("❌ Failed to sign JWT: %v", err)
		return ""
	}
	return signed
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("❌ Failed to write JSON response: %v", err)
	}
}
