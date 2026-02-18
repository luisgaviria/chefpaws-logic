// models/calculator.go
package models

type DogStats struct {
    Weight       float64 `json:"weight"`
    ActivityLevel string  `json:"activity_level"` // e.g., "active", "sedentary"
    Age          int     `json:"age"`
}

type CalculationResult struct {
    DailyCalories float64 `json:"daily_calories"`
    GramsPerMeal  float64 `json:"grams_per_meal"`
    RecipeName    string  `json:"recipe_name"`
}