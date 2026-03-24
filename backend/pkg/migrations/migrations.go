package migrations

import (
	"log"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"gorm.io/gorm"
)

// RunMigrations runs all database migrations
func RunMigrations(database *gorm.DB) error {
	log.Println("Running database migrations...")

	// Auto-migrate the User model
	if err := database.AutoMigrate(&models.User{}); err != nil {
		log.Printf("Error migrating User model: %v", err)
		return err
	}

	log.Println("Migrations completed successfully")
	return nil
}
