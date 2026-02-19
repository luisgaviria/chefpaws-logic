package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os" // Added to handle environment variables

	"github.com/luisgaviria/chefpaws-logic/calculations" // Fixed: Points to new public folder
	"github.com/luisgaviria/chefpaws-logic/internal/api"
	"github.com/luisgaviria/chefpaws-logic/internal/models"
)

func main() {
	// 1. Dynamic Base URL: Uses Railway URL if available, otherwise defaults to DDEV
	baseURL := os.Getenv("DRUPAL_URL")
	if baseURL == "" {
		baseURL = "http://chefpaws-backend.ddev.site"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 1. Homepage Endpoint
	http.HandleFunc("/homepage", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		data, err := api.FetchHomepageData(baseURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(data)
	})

	// 2. Meal Plan Endpoint
	http.HandleFunc("/meal-plan", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		targetID := r.URL.Query().Get("id")

		recipes, _ := api.FetchRecipes(baseURL)
		dogs, _ := api.FetchDogs(baseURL)

		var reports []models.DogReport
		for _, dog := range dogs {
			if targetID != "" && dog.ID != targetID {
				continue
			}

			// Fixed: Changed 'engine' to 'calculations'
			dailyCals := calculations.CalculateDailyCalories(dog)
			portions := make(map[string]float64)

			for _, recipe := range recipes {
				// Fixed: Changed 'engine' to 'calculations'
				portions[recipe.Title] = calculations.CalculatePortionGrams(dailyCals, recipe)
			}

			reports = append(reports, models.DogReport{
				DogName:       dog.Name,
				WeightKG:      dog.WeightKG,
				ActivityLevel: dog.ActivityLevel,
				DailyCalories: math.Round(dailyCals),
				Portions:      portions,
			})
		}
		json.NewEncoder(w).Encode(reports)
	})

	http.HandleFunc("/api/calculate", api.NutritionRevealHandler)
	http.HandleFunc("/api/lead", api.LeadHandler)

	fmt.Printf("🚀 ChefPaws Logic API live at http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}