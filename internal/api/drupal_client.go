package api

import (
	"chefpaws-logic/internal/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type drupalRecipeDTO struct {
	Data []struct {
		ID         string `json:"id"`
		Attributes struct {
			Title        string      `json:"title"`
			Ingredients  string      `json:"field_ingredients"`
			Instructions string      `json:"field_instructions"`
			// We use interface{} to capture whatever Drupal throws at us
			CaloriesRaw  interface{} `json:"field_calories_per_gram"` 
		} `json:"attributes"`
	} `json:"data"`
}

func FetchRecipes(baseURL string) ([]models.Recipe, error) {
	url := fmt.Sprintf("%s/jsonapi/node/recipe", baseURL)
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var dto drupalRecipeDTO
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return nil, err
	}

	var recipes []models.Recipe
	for _, item := range dto.Data {
		// --- DEBUG BLOCK START ---
		// This will print the exact type and value Go is seeing for every recipe
		fmt.Printf("DEBUG [%s]: Raw Type: %T, Value: %+v\n", item.Attributes.Title, item.Attributes.CaloriesRaw, item.Attributes.CaloriesRaw)
		// --- DEBUG BLOCK END ---

		var calories float64

		// Bulletproof conversion: Convert the interface to a string first, then parse
		// This handles float64, string, and even nested maps if Drupal is being tricky
		strVal := fmt.Sprintf("%v", item.Attributes.CaloriesRaw)
		
		// If Drupal sends a nested object, strVal might look like "map[value:1.5]"
		// We'll try to parse the string directly first
		parsed, err := strconv.ParseFloat(strVal, 64)
		if err == nil {
			calories = parsed
		} else {
			// Fallback: Check if it's a map (nested value)
			if m, ok := item.Attributes.CaloriesRaw.(map[string]interface{}); ok {
				if val, exists := m["value"]; exists {
					strVal = fmt.Sprintf("%v", val)
					calories, _ = strconv.ParseFloat(strVal, 64)
				}
			}
		}

		recipes = append(recipes, models.Recipe{
			ID:           item.ID,
			Title:        item.Attributes.Title,
			Ingredients:  item.Attributes.Ingredients,
			Instructions: item.Attributes.Instructions,
			CaloriesPerG: calories,
		})
	}

	return recipes, nil
}

type drupalDogDTO struct {
	Data []struct {
		ID         string `json:"id"`
		Attributes struct {
			Title         string      `json:"title"`
			WeightRaw     interface{} `json:"field_weight_kg"`
			ActivityLevel int         `json:"field_activity_level"`
		} `json:"attributes"`
	} `json:"data"`
}

func FetchDogs(baseURL string) ([]models.DogProfile, error) {
	url := fmt.Sprintf("%s/jsonapi/node/dog", baseURL)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var dto drupalDogDTO
	json.NewDecoder(resp.Body).Decode(&dto)

	var dogs []models.DogProfile
	for _, item := range dto.Data {
		// Convert Weight (interface/string) to float64
		weightStr := fmt.Sprintf("%v", item.Attributes.WeightRaw)
		weight, _ := strconv.ParseFloat(weightStr, 64)

		dogs = append(dogs, models.DogProfile{
			ID:            item.ID,
			Name:          item.Attributes.Title,
			WeightKG:      weight,
			ActivityLevel: item.Attributes.ActivityLevel,
		})
	}
	return dogs, nil
}