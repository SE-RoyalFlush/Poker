package migrations

import (
	"log"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"gorm.io/gorm"
)

// RunMigrations runs all database migrations
func RunMigrations(database *gorm.DB) error {
	log.Println("Running database migrations...")

	// Auto-migrate all persisted models.
	if err := database.AutoMigrate(models.AllModels()...); err != nil {
		log.Printf("Error migrating database models: %v", err)
		return err
	}

	log.Println("Migrations completed successfully")
	return nil
}
