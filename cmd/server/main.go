package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"

	"chefpaws-logic/internal/api"
	"chefpaws-logic/internal/engine"
	"chefpaws-logic/internal/models"
)

type NutritionRequest struct {
	DogName  string  `json:"dogName"`
	Weight   float64 `json:"weight"`
	Activity float64 `json:"activity"` // Multiplier: 1.2 (low) to 2.0 (high)
}

type NutritionResponse struct {
	DogName         string  `json:"dogName"`
	DailyCalories   int     `json:"calories"`
	ProteinTarget   float64 `json:"proteinGrams"`
	RecommendedMix  string  `json:"mix"`
}

func main() {
	baseURL := "http://chefpaws-backend.ddev.site"
	port := ":8080"

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
			if targetID != "" && dog.ID != targetID { continue }

			dailyCals := engine.CalculateDailyCalories(dog)
			portions := make(map[string]float64)

			for _, recipe := range recipes {
				portions[recipe.Title] = engine.CalculatePortionGrams(dailyCals, recipe)
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

	fmt.Printf("🚀 ChefPaws Logic API live at http://localhost%s\n", port)
	log.Fatal(http.ListenAndServe(port, nil))
}