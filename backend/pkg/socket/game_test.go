package socket

import (
	"path/filepath"
	"testing"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/protocol"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/room"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
