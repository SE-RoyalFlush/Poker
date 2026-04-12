package room

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCreateWithGeneratedCodeRetriesOnCollision(t *testing.T) {
	database := setupTestDB(t)
	host := createHostUser(t, database, "host-collision")

	existing, err := Create(database, CreateParams{
		Code:       "AB12CD",
		HostUserID: host.ID,
		MaxPlayers: 6,
	})
	if err != nil {
		t.Fatalf("Create returned error while seeding collision: %v", err)
	}

	room := &models.Room{
		HostUserID: host.ID,
		Status:     models.RoomStatusOpen,
		MaxPlayers: 6,
	}

	attempts := 0
	created, err := createWithGeneratedCode(database, room, func() (string, error) {
		attempts++
		if attempts == 1 {
			return existing.Code, nil
		}
		return "ZX98QP", nil
	})
	if err != nil {
		t.Fatalf("createWithGeneratedCode returned error: %v", err)
	}

	if attempts != 2 {
		t.Fatalf("expected 2 generation attempts, got %d", attempts)
	}
	if created.Code != "ZX98QP" {
		t.Fatalf("expected retry to use ZX98QP, got %q", created.Code)
	}
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db.ResetForTesting()
	t.Cleanup(func() {
		_ = db.Close()
	})

	cfg := &db.Config{
		DatabasePath:    filepath.Join(t.TempDir(), "room-service-internal.db"),
		MaxOpenConns:    10,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Minute,
		LogLevel:        logger.Silent,
	}

	database, err := db.Connect(cfg)
	if err != nil {
		t.Fatalf("failed to connect test database: %v", err)
	}

	return database
}

func createHostUser(t *testing.T, database *gorm.DB, username string) *models.User {
	t.Helper()

	host := &models.User{
		Username:     username,
		PasswordHash: "hash",
	}
	if err := database.Create(host).Error; err != nil {
		t.Fatalf("failed to create host user: %v", err)
	}

	return host
}

func TestCreateWithGeneratedCodeFailsAfterExhaustingRetries(t *testing.T) {
	database := setupTestDB(t)
	host := createHostUser(t, database, "host-collision-exhausted")

	_, err := Create(database, CreateParams{
		Code:       "AB12CD",
		HostUserID: host.ID,
		MaxPlayers: 6,
	})
	if err != nil {
		t.Fatalf("Create returned error while seeding collision: %v", err)
	}

	room := &models.Room{
		HostUserID: host.ID,
		Status:     models.RoomStatusOpen,
		MaxPlayers: 6,
	}

	_, err = createWithGeneratedCode(database, room, func() (string, error) {
		return "AB12CD", nil
	})
	if !errors.Is(err, ErrFailedToGenerateCode) {
		t.Fatalf("expected ErrFailedToGenerateCode, got %v", err)
	}
}
