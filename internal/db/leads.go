package db

import (
	"database/sql"

	"github.com/luisgaviria/chefpaws-logic/internal/models"
)

// SaveLead inserts a new lead record into the MySQL database.
func SaveLead(db *sql.DB, lead models.Lead) error {
	_, err := db.Exec(`
		INSERT INTO leads (owner_name, email, phone, zip, dog_name, daily_calories, portion_grams, special_reqs, source)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		lead.OwnerName,
		lead.Email,
		lead.Phone,
		lead.Zip,
		lead.DogName,
		lead.DailyCalories,
		lead.PortionGrams,
		lead.SpecialReqs,
		lead.Source,
	)
	return err
}
