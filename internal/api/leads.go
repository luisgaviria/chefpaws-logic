package api // or package main, whichever matches your project

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// LeadRequest matches the Svelte JSON structure
type LeadRequest struct {
    DogName        string  `json:"dogName"`
    DailyCalories  float64 `json:"dailyCalories"`
    PortionGrams   float64 `json:"portionGrams"`
    Phone          string  `json:"phone"`
    Zip            string  `json:"zip"`
    Source         string  `json:"source"`
}

// LeadHandler handles the incoming form from Svelte
func LeadHandler(w http.ResponseWriter, r *http.Request) {
    // CORS Headers
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }

    var lead LeadRequest
    if err := json.NewDecoder(r.Body).Decode(&lead); err != nil {
        http.Error(w, "Invalid data", http.StatusBadRequest)
        return
    }

    // This prints to your terminal so you can see it's working!
    fmt.Printf("🎯 NEW LEAD: %s's owner at %s\n", lead.DogName, lead.Phone)

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}