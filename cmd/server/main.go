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
	baseURL := os.Getenv("DRUPAL_URL")
	if baseURL == "" {
		baseURL = "http://chefpaws-backend.ddev.site"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// CHANGED: Now handles any landing page via a "slug" query parameter
	// Example: /page?slug=/home or /page?slug=/our-story
	http.HandleFunc("/page", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		slug := r.URL.Query().Get("slug")
		if slug == "" {
			slug = "/home" // Fallback default
		}

		data, err := api.FetchPageData(baseURL, slug)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(data)
	})

	// Meal Plan Endpoint (Unchanged)
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

	http.HandleFunc("/api/calculate", api.NutritionRevealHandler)
	http.HandleFunc("/api/lead", api.LeadHandler)

	fmt.Printf("🚀 ChefPaws Logic API live at http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}