package models

import "time"

// Lead represents a prospective customer captured via the nutrition calculator.
type Lead struct {
	ID            int64     `json:"id"`
	OwnerName     string    `json:"ownerName"`
	Email         string    `json:"email"`
	Phone         string    `json:"phone"`
	Zip           string    `json:"zip"`
	DogName       string    `json:"dogName"`
	DailyCalories float64   `json:"dailyCalories"`
	PortionGrams  float64   `json:"portionGrams"`
	SpecialReqs   string    `json:"specialRequirements"`
	Source        string    `json:"source"`
	CreatedAt     time.Time `json:"createdAt"`
}
