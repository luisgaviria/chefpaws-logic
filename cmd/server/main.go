package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/luisgaviria/chefpaws-logic/calculations"
	"github.com/luisgaviria/chefpaws-logic/internal/api"
	"github.com/luisgaviria/chefpaws-logic/internal/db"
	"github.com/luisgaviria/chefpaws-logic/internal/models"
)

func main() {
	// Load .env for local development (no-op in production where env vars are injected)
	_ = godotenv.Load()

	// 1. Drupal base URL
	baseURL := os.Getenv("DRUPAL_URL")
	if baseURL == "" {
		baseURL = "http://chefpaws-backend.ddev.site"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 2. MySQL — connect and ensure leads table exists
	var database *sql.DB
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		d, err := db.Connect(dbURL)
		if err != nil {
			log.Printf("⚠️  MySQL connection failed (leads will not be persisted): %v", err)
		} else {
			if err := db.CreateLeadsTable(d); err != nil {
				log.Printf("⚠️  Could not create leads table: %v", err)
			} else {
				fmt.Println("✅ MySQL connected — leads table ready")
			}
			database = d
		}
	} else {
		fmt.Println("⚠️  DATABASE_URL not set — leads will not be persisted")
	}

	resendKey := os.Getenv("RESEND_API_KEY")

	// 3. Page handler
	http.HandleFunc("/page", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		slug := r.URL.Query().Get("slug")
		if slug == "" {
			slug = "/"
		}

		data, err := api.FetchPageData(baseURL, slug)
		if err != nil {
			fmt.Printf("❌ Error fetching page data for %s: %v\n", slug, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(data)
	})

	// 4. Meal plan endpoint
	http.HandleFunc("/meal-plan", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		targetID := r.URL.Query().Get("id")

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

	// 5. Utility handlers
	http.HandleFunc("/api/calculate", api.NutritionRevealHandler)
	http.HandleFunc("/api/lead", api.NewLeadHandler(database, resendKey))

	fmt.Printf("🚀 ChefPaws Logic API live at http://localhost:%s\n", port)
	fmt.Printf("🔗 Connecting to Drupal at: %s\n", baseURL)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
