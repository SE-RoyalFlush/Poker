package room_test

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/room"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCreatePersistsRoomAndHostAssociation(t *testing.T) {
	database := setupTestDB(t)
	host := createHostUser(t, database, "host-create")

	createdRoom, err := room.Create(database, room.CreateParams{
		Code:       "ab12cd",
		HostUserID: host.ID,
		Status:     models.RoomStatusOpen,
		MaxPlayers: 8,
		IsPrivate:  true,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	var persisted models.Room
	if err := database.First(&persisted, createdRoom.ID).Error; err != nil {
		t.Fatalf("failed to reload persisted room: %v", err)
	}

	if persisted.Code != "AB12CD" {
		t.Fatalf("expected normalized room code AB12CD, got %q", persisted.Code)
	}
	if persisted.HostUserID != host.ID {
		t.Fatalf("expected HostUserID %d, got %d", host.ID, persisted.HostUserID)
	}
	if persisted.Status != models.RoomStatusOpen {
		t.Fatalf("expected status %q, got %q", models.RoomStatusOpen, persisted.Status)
	}
	if persisted.MaxPlayers != 8 {
		t.Fatalf("expected MaxPlayers 8, got %d", persisted.MaxPlayers)
	}
	if !persisted.IsPrivate {
		t.Fatal("expected IsPrivate to be true")
	}
}

func TestFindByCodeReturnsMatchingRoom(t *testing.T) {
	database := setupTestDB(t)
	host := createHostUser(t, database, "host-lookup")

	expected, err := room.Create(database, room.CreateParams{
		Code:       "zx98qp",
		HostUserID: host.ID,
		Status:     models.RoomStatusInGame,
		MaxPlayers: 6,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	found, err := room.FindByCode(database, "  zx98qp ")
	if err != nil {
		t.Fatalf("FindByCode returned error: %v", err)
	}

	if found.ID != expected.ID {
		t.Fatalf("expected room ID %d, got %d", expected.ID, found.ID)
	}
	if found.Code != "ZX98QP" {
		t.Fatalf("expected code ZX98QP, got %q", found.Code)
	}
	if found.HostUserID != host.ID {
		t.Fatalf("expected HostUserID %d, got %d", host.ID, found.HostUserID)
	}
	if found.HostUser.ID != host.ID {
		t.Fatalf("expected preloaded host ID %d, got %d", host.ID, found.HostUser.ID)
	}
}

func TestCreateRejectsInvalidRoomCodes(t *testing.T) {
	database := setupTestDB(t)
	host := createHostUser(t, database, "host-invalid-create")

	testCases := []struct {
		name string
		code string
	}{
		{name: "blank", code: "   "},
		{name: "too short", code: "ABC12"},
		{name: "too long", code: "ABC1234"},
		{name: "invalid characters", code: "AB-12!"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			createdRoom, err := room.Create(database, room.CreateParams{
				Code:       tc.code,
				HostUserID: host.ID,
			})
			if !errors.Is(err, room.ErrInvalidRoomCode) {
				t.Fatalf("expected ErrInvalidRoomCode, got %v", err)
			}
			if createdRoom != nil {
				t.Fatalf("expected no room to be created, got %+v", createdRoom)
			}
		})
	}
}

func TestFindByCodeRejectsInvalidRoomCodes(t *testing.T) {
	database := setupTestDB(t)

	found, err := room.FindByCode(database, " ")
	if !errors.Is(err, room.ErrInvalidRoomCode) {
		t.Fatalf("expected ErrInvalidRoomCode, got %v", err)
	}
	if found != nil {
		t.Fatalf("expected no room result, got %+v", found)
	}
}

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db.ResetForTesting()
	t.Cleanup(func() {
		_ = db.Close()
	})

	cfg := &db.Config{
		DatabasePath:    filepath.Join(t.TempDir(), "room-service.db"),
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
