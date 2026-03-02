package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/luisgaviria/chefpaws-logic/internal/db"
	"github.com/luisgaviria/chefpaws-logic/internal/email"
	"github.com/luisgaviria/chefpaws-logic/internal/models"
)

// LeadRequest matches the JSON payload sent by the Svelte nutrition calculator.
type LeadRequest struct {
	OwnerName     string  `json:"ownerName"`
	Email         string  `json:"email"`
	Phone         string  `json:"phone"`
	Zip           string  `json:"zip"`
	DogName       string  `json:"dogName"`
	DailyCalories float64 `json:"dailyCalories"`
	PortionGrams  float64 `json:"portionGrams"`
	SpecialReqs   string  `json:"specialRequirements"`
	Source        string  `json:"source"`
}

// NewLeadHandler returns an http.HandlerFunc wired to MySQL and Resend.
// Passing a nil database or empty resendKey skips those steps gracefully.
func NewLeadHandler(database *sql.DB, resendKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		var req LeadRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		lead := models.Lead{
			OwnerName:     req.OwnerName,
			Email:         req.Email,
			Phone:         req.Phone,
			Zip:           req.Zip,
			DogName:       req.DogName,
			DailyCalories: req.DailyCalories,
			PortionGrams:  req.PortionGrams,
			SpecialReqs:   req.SpecialReqs,
			Source:        req.Source,
		}

		fmt.Printf("🎯 NEW LEAD: %s's owner %s <%s> phone=%s zip=%s\n",
			lead.DogName, lead.OwnerName, lead.Email, lead.Phone, lead.Zip)

		if database != nil {
			if err := db.SaveLead(database, lead); err != nil {
				fmt.Printf("⚠️  MySQL save failed: %v\n", err)
			} else {
				fmt.Printf("✅ Lead saved to MySQL: %s\n", lead.Email)
			}
		}

		if resendKey != "" {
			if err := email.SendLeadNotification(resendKey, lead); err != nil {
				fmt.Printf("⚠️  Resend failed: %v\n", err)
			} else {
				fmt.Printf("📧 Notification sent to %s\n", lead.Email)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}
}
