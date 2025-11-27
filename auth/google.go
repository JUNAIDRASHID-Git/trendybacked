package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"gorm.io/gorm"
)

// ---------------------------------------------
// GOOGLE USER LOGIN
// ---------------------------------------------
func GoogleUserLoginHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	var req struct {
		IDToken string `json:"idToken"`
		GuestID string `json:"guest_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.IDToken == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Verify Firebase token
	token, err := firebaseAuth.VerifyIDTokenAndCheckRevoked(ctx, req.IDToken)
	if err != nil {
		http.Error(w, "Invalid Firebase ID token", http.StatusUnauthorized)
		return
	}

	if token.Audience != projectID {
		http.Error(w, "Invalid token audience", http.StatusUnauthorized)
		return
	}

	// Extract user info
	email := token.Claims["email"].(string)
	name, _ := token.Claims["name"].(string)
	picture, _ := token.Claims["picture"].(string)
	firebaseUserID := token.UID

	// ---------------------------------------------
	// 1️⃣ Fetch or Create user
	// ---------------------------------------------
	var user models.User
	err = db.Preload("Cart.Items").Where("id = ?", firebaseUserID).First(&user).Error

	if err == gorm.ErrRecordNotFound {
		// User does not exist → Create
		user = models.User{
			ID:       firebaseUserID,
			Email:    email,
			Name:     name,
			Picture:  picture,
			Provider: "google",
			Cart:     models.Cart{UserID: firebaseUserID},
		}

		if err := db.Create(&user).Error; err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

	} else if err == nil {
		// User already exists → Update profile
		db.Model(&user).Updates(models.User{
			Name:    name,
			Picture: picture,
		})
	} else {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// ---------------------------------------------
	// 2️⃣ Merge Guest Cart → User Cart
	// ---------------------------------------------
	var mergeStatus string = "no-guest-cart"

	if req.GuestID != "" {
		merged, err := mergeGuestCartIntoUserCart(db, req.GuestID, user.ID)
		if err != nil {
			mergeStatus = "merge-failed"
		} else if merged {
			mergeStatus = "merged-success"
		} else {
			mergeStatus = "guest-cart-empty"
		}
	}

	// ---------------------------------------------
	// 3️⃣ Create auth response
	// ---------------------------------------------
	resp := map[string]interface{}{
		"message":         "Login successful",
		"merge_status":    mergeStatus,
		"user":            user,
		"firebase_id":     firebaseUserID,
		"profile_updated": true,
		"token":           issueJWT(email, "user", firebaseUserID, name, picture),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ---------------------------------------------
// MERGE GUEST CART INTO USER CART
// RETURNS: (bool merged, error)
// ---------------------------------------------
func mergeGuestCartIntoUserCart(db *gorm.DB, guestID, userID string) (bool, error) {
	tx := db.Begin()
	if tx.Error != nil {
		return false, tx.Error
	}

	// --------------------
	// Load guest cart
	// --------------------
	var guestCart models.GuestCart
	if err := tx.Preload("Items").
		Where("guest_id = ?", guestID).
		First(&guestCart).Error; err != nil {

		tx.Rollback()
		return false, nil // nothing to merge
	}

	// --------------------
	// Load or create user cart
	// --------------------
	var userCart models.Cart
	err := tx.Preload("Items").
		Where("user_id = ?", userID).
		First(&userCart).Error

	if err == gorm.ErrRecordNotFound {
		userCart = models.Cart{UserID: userID}
		if err := tx.Create(&userCart).Error; err != nil {
			tx.Rollback()
			return false, err
		}

		tx.Preload("Items").Where("user_id = ?", userID).First(&userCart)
	} else if err != nil {
		tx.Rollback()
		return false, err
	}

	// --------------------
	// Merge items
	// --------------------
	for _, guestItem := range guestCart.Items {
		var userItem models.CartItem

		lookupErr := tx.Where(
			"cart_id = ? AND product_id = ?",
			userCart.CartID,
			guestItem.ProductID,
		).First(&userItem).Error

		if lookupErr == nil {
			// Update quantity
			userItem.Quantity += guestItem.Quantity
			userItem.AddedAt = time.Now()

			if err := tx.Save(&userItem).Error; err != nil {
				tx.Rollback()
				return false, err
			}

		} else if lookupErr == gorm.ErrRecordNotFound {
			// Insert new item
			newItem := models.CartItem{
				CartID:              userCart.CartID,
				ProductID:           guestItem.ProductID,
				ProductEName:        guestItem.ProductEName,
				ProductArName:       guestItem.ProductArName,
				ProductImage:        guestItem.ProductImage,
				ProductStock:        guestItem.ProductStock,
				ProductSalePrice:    guestItem.ProductSalePrice,
				ProductRegularPrice: guestItem.ProductRegularPrice,
				Weight:              guestItem.Weight,
				Quantity:            guestItem.Quantity,
				AddedAt:             time.Now(),
			}

			if err := tx.Create(&newItem).Error; err != nil {
				tx.Rollback()
				return false, err
			}

		} else {
			tx.Rollback()
			return false, lookupErr
		}
	}

	// --------------------
	// Delete guest cart
	// --------------------
	if err := tx.Where("cart_id = ?", guestCart.CartID).Delete(&models.GuestCartItem{}).Error; err != nil {
		tx.Rollback()
		return false, err
	}

	if err := tx.Delete(&guestCart).Error; err != nil {
		tx.Rollback()
		return false, err
	}

	// Commit
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return false, err
	}

	return true, nil
}

// issueJWT generates a JWT token for a user
func issueJWT(email, role, userID, name, picture string) string {
	claims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"role":    role,
		"name":    name,
		"picture": picture,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		// In production, you may want to handle this differently
		return ""
	}

	return signedToken
}
