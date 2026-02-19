package calculations

import (
	"math"

	// Using the full path ensures the cloud can find your models
	"github.com/luisgaviria/chefpaws-logic/internal/models"
)

// CalculateDailyCalories determines how much a dog needs to eat.
func CalculateDailyCalories(dog models.DogProfile) float64 {
	// RER = 70 * (weight_kg ^ 0.75)
	rer := 70 * math.Pow(dog.WeightKG, 0.75)

	// Adjust based on activity level (1: Couch Potato, 5: Athlete)
	// Typical maintenance is 1.6x RER
	activityMultiplier := 1.0 + (float64(dog.ActivityLevel) * 0.2)

	return rer * activityMultiplier
}

// CalculatePortionGrams returns the amount of food in grams.
func CalculatePortionGrams(DailyCalories float64, recipe models.Recipe) float64 {
	if recipe.CaloriesPerG == 0 {
		return 0 // Avoid division by zero
	}
	grams := DailyCalories / recipe.CaloriesPerG
	return math.Round(grams)
}