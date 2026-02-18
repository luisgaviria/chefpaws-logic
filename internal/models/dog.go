package models

type DogProfile struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	WeightKG      float64 `json:"weight_kg"`
	ActivityLevel int     `json:"activity_level"`
}

//DogReport structure defines what the Website will receive
type DogReport struct {
	DogName       string             `json:"dog_name"`
	WeightKG      float64            `json:"weight_kg"`
	ActivityLevel int                `json:"activity_level"`
	DailyCalories float64            `json:"daily_calories"`
	Portions      map[string]float64 `json:"portions"`
}
