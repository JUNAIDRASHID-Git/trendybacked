// GetAllAdminsHandler fetches all admin users from the database.
package controllers

import (
	"encoding/json"
	"github.com/junaidrashid-git/ecommerce-api/models"
	"gorm.io/gorm"
	"log"
	"net/http"
)

func GetAllAdminsHandler(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	var admins []models.Admin

	// Fetch all admins from the database
	if err := db.Find(&admins).Error; err != nil {
		log.Println("❌ Failed to fetch admins:", err)
		http.Error(w, "Failed to fetch admins", http.StatusInternalServerError)
		return
	}

	// Respond with the list of admins
	respondJSON(w, http.StatusOK, admins)
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
