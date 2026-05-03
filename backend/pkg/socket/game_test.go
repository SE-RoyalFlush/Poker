package socket

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/game"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/protocol"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/room"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestHandleGameActionFoldBroadcastsGameOver verifies that when a player folds
// during an active game, GAME_OVER is broadcast to all room clients and the
// result is persisted.
func TestHandleGameActionFoldBroadcastsGameOver(t *testing.T) {
	database := setupCompleteHandTestDB(t)
	winner := models.User{Username: "winner-ga", PasswordHash: "hash"}
	folder := models.User{Username: "folder-ga", PasswordHash: "hash"}
	if err := database.Create(&winner).Error; err != nil {
		t.Fatalf("failed to create winner: %v", err)
	}
	if err := database.Create(&folder).Error; err != nil {
		t.Fatalf("failed to create folder: %v", err)
	}

	hub := NewHub()
	winnerClient := &Client{
		hub:      hub,
		user:     &winner,
		send:     make(chan Message, 10),
		roomCode: "GAFLD1",
	}
	folderClient := &Client{
		hub:      hub,
		user:     &folder,
		send:     make(chan Message, 10),
		roomCode: "GAFLD1",
	}

	lobby := room.NewLobby()
	lobby.JoinPlayer(room.Player{ID: winner.ID, Username: winner.Username})
	lobby.JoinPlayer(room.Player{ID: folder.ID, Username: folder.Username})

	// Start a game manually and inject it into the hub.
	g := game.NewGame(lobby.Players(), "GAFLD1")
	events := g.Start()

	clients := []*Client{winnerClient, folderClient}
	hub.rooms["GAFLD1"] = &roomState{
		lobby:   lobby,
		clients: map[*Client]bool{winnerClient: true, folderClient: true},
		game:    g,
	}

	// Drain start events.
	hub.dispatchGameEvents("GAFLD1", clients, events)
	for _, c := range clients {
		for len(c.send) > 0 {
			<-c.send
		}
	}

	// Inject the db singleton so persistResult works.
	db.ResetForTesting()
	t.Cleanup(func() { _ = db.Close() })
	tmpDB, err := db.Connect(&db.Config{
		DatabasePath:    filepath.Join(t.TempDir(), "ga-fold-test.db"),
		MaxOpenConns:    5,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Minute,
		LogLevel:        logger.Silent,
	})
	if err != nil {
		t.Fatalf("db connect: %v", err)
	}
	if err := tmpDB.AutoMigrate(&models.User{}, &models.GameResult{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := tmpDB.Create(&winner).Error; err != nil {
		t.Logf("note: could not seed winner in tmp db (might already exist): %v", err)
	}

	// The active player folds (whoever is first to act).
	activeID := g.ActivePlayerID()
	var activeClient *Client
	for _, c := range clients {
		if c.user.ID == activeID {
			activeClient = c
			break
		}
	}
	if activeClient == nil {
		t.Fatal("could not find active client")
	}

	hub.HandleGameAction(activeClient, game.ActionFold, 0)

	// Both clients receive PLAYER_ACTION then GAME_OVER.
	for _, c := range clients {
		var gameOverMsg *Message
		for i := 0; i < 5; i++ {
			select {
			case msg := <-c.send:
				if msg.Type == protocol.ServerMsgGameOver {
					m := msg
					gameOverMsg = &m
				}
				// skip PLAYER_ACTION and other intermediate messages
			default:
				i = 5 // break
			}
		}
		if gameOverMsg == nil {
			t.Fatalf("client %s did not receive GAME_OVER", c.user.Username)
		}
	}
}

func TestHandleGameActionErrorWhenNoGameActive(t *testing.T) {
	hub := NewHub()
	client := &Client{
		hub:      hub,
		user:     &models.User{Username: "no-game-player"},
		send:     make(chan Message, 2),
		roomCode: "NOGAME",
	}
	hub.rooms["NOGAME"] = &roomState{
		lobby:   room.NewLobby(),
		clients: map[*Client]bool{client: true},
		game:    nil,
	}

	hub.HandleGameAction(client, game.ActionFold, 0)

	select {
	case msg := <-client.send:
		if msg.Type != protocol.ServerMsgError {
			t.Fatalf("expected ERROR, got %q", msg.Type)
		}
	default:
		t.Fatal("expected ERROR message when no game is active")
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
