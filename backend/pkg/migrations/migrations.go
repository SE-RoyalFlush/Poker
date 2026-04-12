package migrations

import (
	"log"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"gorm.io/gorm"
)

// RunMigrations runs all database migrations
func RunMigrations(database *gorm.DB) error {
	log.Println("Running database migrations...")

	if err := database.AutoMigrate(models.AllModels()...); err != nil {
		log.Printf("Error running model migrations: %v", err)
		return err
	}

	log.Println("Migrations completed successfully")
	return nil
}
