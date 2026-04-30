package game

import (
	"errors"
	"time"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"gorm.io/gorm"
)

var (
	ErrMissingDatabase = errors.New("database is required")
	ErrMissingWinner   = errors.New("winner ID is required")
	ErrNegativePot     = errors.New("pot size cannot be negative")
)

// SaveGameResult persists a completed hand result using GORM.
func SaveGameResult(database *gorm.DB, result models.GameResult) error {
	if database == nil {
		return ErrMissingDatabase
	}
	if result.WinnerID == 0 {
		return ErrMissingWinner
	}
	if result.PotSize < 0 {
		return ErrNegativePot
	}
	if result.Date.IsZero() {
		result.Date = time.Now().UTC()
	}
	if result.GameType == "" {
		result.GameType = models.DefaultGameType
	}

	return database.Create(&result).Error
}
