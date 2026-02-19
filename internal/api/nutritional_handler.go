package api

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"

	"github.com/luisgaviria/chefpaws-logic/calculations"
	"github.com/luisgaviria/chefpaws-logic/internal/models"
)

// DrupalCollection maps to the /jsonapi/node/recipe list
type DrupalCollection struct {
	Data []struct {
		Attributes struct {
			Title           string `json:"title"`
			Ingredients     string `json:"field_ingredients"`
			CaloriesPerGram string `json:"field_calories_per_gram"`
		} `json:"attributes"`
	} `json:"data"`
}

func NutritionRevealHandler(w http.ResponseWriter, r *http.Request) {
	// 1. CORS Headers
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4321")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 2. Decode Input
	var req struct {
		Name          string  `json:"name"`
		WeightKG      float64 `json:"weightKG"`
		ActivityLevel int     `json:"activityLevel"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// 3. Calculation
	dog := models.DogProfile{
		Name: req.Name, WeightKG: req.WeightKG, ActivityLevel: req.ActivityLevel,
	}
	dailyCals := calculations.CalculateDailyCalories(dog)

	// 4. Fetch ALL recipes from Drupal
	drupalURL := "http://chefpaws-backend.ddev.site/jsonapi/node/recipe"
	resp, err := http.Get(drupalURL)
	
	var recipes []map[string]interface{}

	if err == nil && resp.StatusCode == http.StatusOK {
		var collection DrupalCollection
		if err := json.NewDecoder(resp.Body).Decode(&collection); err == nil {
			for _, item := range collection.Data {
				attr := item.Attributes
				cals, _ := strconv.ParseFloat(attr.CaloriesPerGram, 64)
				
				// Calculate portion specifically for this recipe's density
				portion := dailyCals / cals

				recipes = append(recipes, map[string]interface{}{
					"recipeName":   attr.Title,
					"ingredients":  attr.Ingredients,
					"portionGrams": math.Round(portion),
				})
			}
		}
		resp.Body.Close()
	}

	// 5. Send the list back to Astro
	response := map[string]interface{}{
		"dogName":       dog.Name,
		"dailyCalories": math.Round(dailyCals),
		"recipes":       recipes,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}