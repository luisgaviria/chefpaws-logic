package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"

	"github.com/luisgaviria/chefpaws-logic/calculations"
	"github.com/luisgaviria/chefpaws-logic/internal/api"
	"github.com/luisgaviria/chefpaws-logic/internal/models"
)

func main() {
	// 1. Environment-aware Base URL
	baseURL := os.Getenv("DRUPAL_URL")
	if baseURL == "" {
		// Local DDEV default
		baseURL = "http://chefpaws-backend.ddev.site"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 2. PAGE HANDLER
	// Handles landing pages. Defaults to "/" to trigger the UUID-based homepage fetch.
	http.HandleFunc("/page", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		slug := r.URL.Query().Get("slug")
		if slug == "" {
			slug = "/" // Changed from "/home" to "/" for UUID compatibility
		}

		data, err := api.FetchPageData(baseURL, slug)
		if err != nil {
			fmt.Printf("❌ Error fetching page data for %s: %v\n", slug, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(data)
	})

	// 3. MEAL PLAN ENDPOINT
	// Calculates personalized nutrition reports for specific dogs.
	http.HandleFunc("/meal-plan", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		targetID := r.URL.Query().Get("id")
		
		// Fetch core data from Drupal
		recipes, err := api.FetchRecipes(baseURL)
		if err != nil {
			fmt.Printf("❌ Error fetching recipes: %v\n", err)
		}
		
		dogs, err := api.FetchDogs(baseURL)
		if err != nil {
			fmt.Printf("❌ Error fetching dogs: %v\n", err)
		}

		var reports []models.DogReport
		for _, dog := range dogs {
			// If an ID is provided, filter for only that dog
			if targetID != "" && dog.ID != targetID {
				continue
			}

			dailyCals := calculations.CalculateDailyCalories(dog)
			portions := make(map[string]float64)

			for _, recipe := range recipes {
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

	// 4. UTILITY HANDLERS
	// Lead generation and specialized nutrition reveal logic
	http.HandleFunc("/api/calculate", api.NutritionRevealHandler)
	http.HandleFunc("/api/lead", api.LeadHandler)

	fmt.Printf("🚀 ChefPaws Logic API live at http://localhost:%s\n", port)
	fmt.Printf("🔗 Connecting to Drupal at: %s\n", baseURL)
	
	log.Fatal(http.ListenAndServe(":"+port, nil))
}