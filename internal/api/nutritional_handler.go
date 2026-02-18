// internal/api/nutrition_handler.go
package api

import (
	"chefpaws-logic/internal/engine"
	"chefpaws-logic/internal/models"
	"encoding/json"
	"math"
	"net/http"
)

// internal/api/handlers.go

func NutritionRevealHandler(w http.ResponseWriter, r *http.Request) {
	// 1. Set the Allow-Origin to match your Astro dev port
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4321")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// 2. Handle the Preflight (the browser sends this first to check permissions)
	if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
	}

	// 3. Your existing calculation logic...
	var req struct {
			Name     string  `json:"name"`
			WeightKG float64 `json:"weightKG"`
			Activity int     `json:"activityLevel"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
	}

	dog := models.DogProfile{
			Name: req.Name, 
			WeightKG: req.WeightKG, 
			ActivityLevel: req.Activity,
	}
	
	dailyCals := engine.CalculateDailyCalories(dog)
	
	json.NewEncoder(w).Encode(map[string]interface{}{
			"dogName":       dog.Name,
			"dailyCalories": math.Round(dailyCals),
			"portionGrams":  math.Round(dailyCals / 1.5), 
	})
}