package game_test

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/game"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSaveGameResultPersistsCompletedHand(t *testing.T) {
	database := setupGameResultTestDB(t)
	winner := createGameResultTestUser(t, database, "winner")

	result := models.GameResult{
		WinnerID: winner.ID,
		PotSize:  450,
	}
	if err := game.SaveGameResult(database, result); err != nil {
		t.Fatalf("SaveGameResult returned error: %v", err)
	}

	var saved models.GameResult
	if err := database.First(&saved).Error; err != nil {
		t.Fatalf("failed to load saved game result: %v", err)
	}
	if saved.WinnerID != winner.ID {
		t.Fatalf("WinnerID = %d, want %d", saved.WinnerID, winner.ID)
	}
	if saved.PotSize != 450 {
		t.Fatalf("PotSize = %d, want 450", saved.PotSize)
	}
	if saved.GameType != models.DefaultGameType {
		t.Fatalf("GameType = %q, want %q", saved.GameType, models.DefaultGameType)
	}
	if saved.Date.IsZero() {
		t.Fatal("expected Date to be set")
	}
}

func TestSaveGameResultValidatesInput(t *testing.T) {
	database := setupGameResultTestDB(t)

	tests := []struct {
		name   string
		db     *gorm.DB
		result models.GameResult
		want   error
	}{
		{
			name:   "missing database",
			db:     nil,
			result: models.GameResult{WinnerID: 1, PotSize: 10},
			want:   game.ErrMissingDatabase,
		},
		{
			name:   "missing winner",
			db:     database,
			result: models.GameResult{PotSize: 10},
			want:   game.ErrMissingWinner,
		},
		{
			name:   "negative pot",
			db:     database,
			result: models.GameResult{WinnerID: 1, PotSize: -1},
			want:   game.ErrNegativePot,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := game.SaveGameResult(tt.db, tt.result)
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
		})
	}
}

func setupGameResultTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	database, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "game-result-test.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := database.AutoMigrate(&models.User{}, &models.GameResult{}); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return database
}

func createGameResultTestUser(t *testing.T, database *gorm.DB, username string) models.User {
	t.Helper()

	user := models.User{
		Username:     username,
		PasswordHash: "hash",
	}
	if err := database.Create(&user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	return user
}
