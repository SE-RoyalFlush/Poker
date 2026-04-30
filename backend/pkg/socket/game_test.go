package socket

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/protocol"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/room"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestCompleteHandPersistsGameResult(t *testing.T) {
	database := setupCompleteHandTestDB(t)
	winner := models.User{
		Username:     "winner",
		PasswordHash: "hash",
	}
	if err := database.Create(&winner).Error; err != nil {
		t.Fatalf("failed to create winner: %v", err)
	}

	hub := NewHub()
	payload := protocol.GameOverPayload{
		WinnerID: winner.ID,
		Pot:      900,
	}
	hub.CompleteHand(database, "ABC123", payload)

	var result models.GameResult
	if err := database.First(&result).Error; err != nil {
		t.Fatalf("failed to load persisted game result: %v", err)
	}
	if result.WinnerID != winner.ID {
		t.Fatalf("WinnerID = %d, want %d", result.WinnerID, winner.ID)
	}
	if result.PotSize != payload.Pot {
		t.Fatalf("PotSize = %d, want %d", result.PotSize, payload.Pot)
	}
}

func TestCompleteHandBroadcastsGameOverWhenPersistenceFails(t *testing.T) {
	hub := NewHub()
	client := &Client{
		user:     nil,
		send:     make(chan Message, 1),
		roomCode: "ABC123",
	}
	hub.rooms["ABC123"] = &roomState{
		clients: map[*Client]bool{client: true},
	}

	payload := protocol.GameOverPayload{
		WinnerID: 7,
		Pot:      900,
	}
	hub.CompleteHand(nil, "ABC123", payload)

	select {
	case message := <-client.send:
		if message.Type != protocol.ServerMsgGameOver {
			t.Fatalf("message type = %q, want %q", message.Type, protocol.ServerMsgGameOver)
		}
		got, ok := message.Payload.(protocol.GameOverPayload)
		if !ok {
			t.Fatalf("payload type = %T, want protocol.GameOverPayload", message.Payload)
		}
		if got.WinnerID != payload.WinnerID || got.Pot != payload.Pot {
			t.Fatalf("payload = %+v, want %+v", got, payload)
		}
	default:
		t.Fatal("expected GAME_OVER broadcast")
	}
}

func TestCompleteFoldPersistsAndBroadcastsWinner(t *testing.T) {
	database := setupCompleteHandTestDB(t)
	winner := models.User{Username: "winner-fold", PasswordHash: "hash"}
	folder := models.User{Username: "folder-fold", PasswordHash: "hash"}
	if err := database.Create(&winner).Error; err != nil {
		t.Fatalf("failed to create winner: %v", err)
	}
	if err := database.Create(&folder).Error; err != nil {
		t.Fatalf("failed to create folder: %v", err)
	}

	hub := NewHub()
	winnerClient := &Client{
		user:     &winner,
		send:     make(chan Message, 1),
		roomCode: "FOLD01",
	}
	folderClient := &Client{
		user:     &folder,
		send:     make(chan Message, 1),
		roomCode: "FOLD01",
	}
	lobby := room.NewLobby()
	lobby.JoinPlayer(room.Player{ID: winner.ID, Username: winner.Username})
	lobby.JoinPlayer(room.Player{ID: folder.ID, Username: folder.Username})
	hub.rooms["FOLD01"] = &roomState{
		lobby: lobby,
		clients: map[*Client]bool{
			winnerClient: true,
			folderClient: true,
		},
	}

	hub.CompleteFold(database, folderClient)

	var result models.GameResult
	if err := database.First(&result).Error; err != nil {
		t.Fatalf("failed to load persisted fold result: %v", err)
	}
	if result.WinnerID != winner.ID {
		t.Fatalf("WinnerID = %d, want %d", result.WinnerID, winner.ID)
	}

	for _, client := range []*Client{winnerClient, folderClient} {
		select {
		case message := <-client.send:
			if message.Type != protocol.ServerMsgGameOver {
				t.Fatalf("message type = %q, want %q", message.Type, protocol.ServerMsgGameOver)
			}
			payload := message.Payload.(protocol.GameOverPayload)
			if payload.WinnerID != winner.ID {
				t.Fatalf("winner ID = %d, want %d", payload.WinnerID, winner.ID)
			}
		default:
			t.Fatal("expected GAME_OVER broadcast")
		}
	}
}

func TestLeaveRemovesEmptyRoomAndMarksPersistedRoomInactive(t *testing.T) {
	database := setupHubCleanupTestDB(t)
	user := createHubCleanupUser(t, database, "cleanup-host")
	roomModel := createHubCleanupRoom(t, database, "CLEAN1", user.ID)

	hub := NewHub()
	client := &Client{
		user: user,
		send: make(chan Message, 1),
	}
	if _, err := hub.Join(roomModel, client); err != nil {
		t.Fatalf("failed to join room: %v", err)
	}

	hub.Leave(client)

	if got := hub.Occupancy(roomModel.Code); got != 0 {
		t.Fatalf("expected empty room occupancy 0, got %d", got)
	}
	if _, ok := hub.rooms[roomModel.Code]; ok {
		t.Fatal("expected empty room to be removed from hub")
	}
	if got := client.roomCodeValue(); got != "" {
		t.Fatalf("expected client room code to be cleared, got %q", got)
	}

	var persisted models.Room
	if err := database.Where("code = ?", roomModel.Code).First(&persisted).Error; err != nil {
		t.Fatalf("failed to reload room: %v", err)
	}
	if persisted.IsActive {
		t.Fatal("expected persisted room to be marked inactive")
	}
}

func TestRepeatedJoinLeaveCyclesDoNotRetainRooms(t *testing.T) {
	database := setupHubCleanupTestDB(t)
	user := createHubCleanupUser(t, database, "cleanup-cycle")
	hub := NewHub()

	for i, code := range []string{"CYC001", "CYC002", "CYC003"} {
		roomModel := createHubCleanupRoom(t, database, code, user.ID)
		client := &Client{
			user: user,
			send: make(chan Message, 1),
		}
		if _, err := hub.Join(roomModel, client); err != nil {
			t.Fatalf("cycle %d failed to join room: %v", i, err)
		}

		hub.Leave(client)
	}

	if len(hub.rooms) != 0 {
		t.Fatalf("expected no retained rooms, got %d", len(hub.rooms))
	}

	var activeCount int64
	if err := database.Model(&models.Room{}).Where("is_active = ?", true).Count(&activeCount).Error; err != nil {
		t.Fatalf("failed to count active rooms: %v", err)
	}
	if activeCount != 0 {
		t.Fatalf("expected no active rooms after repeated cleanup, got %d", activeCount)
	}
}

func setupCompleteHandTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	database, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "complete-hand-test.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := database.AutoMigrate(&models.User{}, &models.GameResult{}); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return database
}

func setupHubCleanupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	tempDir := t.TempDir()
	db.ResetForTesting()
	t.Cleanup(func() {
		_ = db.Close()
	})

	database, err := db.Connect(&db.Config{
		DatabasePath:    filepath.Join(tempDir, "hub-cleanup-test.db"),
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Minute,
		LogLevel:        logger.Silent,
	})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	return database
}

func createHubCleanupUser(t *testing.T, database *gorm.DB, username string) *models.User {
	t.Helper()

	user := &models.User{
		Username:     username,
		PasswordHash: "hash",
	}
	if err := database.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	return user
}

func createHubCleanupRoom(t *testing.T, database *gorm.DB, code string, hostUserID uint) *models.Room {
	t.Helper()

	roomModel, err := room.Create(database, room.CreateParams{
		Code:       code,
		HostUserID: hostUserID,
		MaxPlayers: 6,
	})
	if err != nil {
		t.Fatalf("failed to create room %s: %v", code, err)
	}

	return roomModel
}
