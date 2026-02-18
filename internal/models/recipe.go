package models

// Recipe represents the core business object for a dog meal.
// This is what the Svelte frontend and the Go engine will use.
type Recipe struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Ingredients  string  `json:"ingredients"`
	Instructions string  `json:"instructions"`
	CaloriesPerG float64 `json:"calories_per_gram"`
}
