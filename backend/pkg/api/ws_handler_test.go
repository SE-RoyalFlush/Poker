package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/auth"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/room"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/net/websocket"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestWebSocketHandlerJoinToggleReadyAndDisconnect(t *testing.T) {
	database := setupWebSocketTestDB(t)
	resetWebSocketStateForTesting()

	host := createWebSocketUser(t, database, "host-user")
	guest := createWebSocketUser(t, database, "guest-user")
	roomModel, err := room.Create(database, room.CreateParams{
		Code:       "AB12CD",
		HostUserID: host.ID,
		MaxPlayers: 6,
	})
	if err != nil {
		t.Fatalf("failed to create room: %v", err)
	}

	server := httptest.NewServer(NewRouter())
	defer server.Close()

	hostConn := dialWebSocket(t, server.URL, sessionCookieForTest(t, "host-user"))
	defer hostConn.Close()
	guestConn := dialWebSocket(t, server.URL, sessionCookieForTest(t, "guest-user"))
	defer guestConn.Close()

	writeWSMessage(t, hostConn, wsMessage{
		Type:    messageTypeJoinRoom,
		Payload: joinRoomPayload{RoomCode: roomModel.Code},
	})

	hostJoined := readWSMessage(t, hostConn)
	if hostJoined.Type != messageTypePlayerJoined {
		t.Fatalf("expected host join message type %q, got %q", messageTypePlayerJoined, hostJoined.Type)
	}
	hostPlayer := decodeRoomPlayer(t, hostJoined.Payload)
	if hostPlayer.Username != "host-user" || !hostPlayer.IsHost || hostPlayer.IsReady {
		t.Fatalf("unexpected host player payload: %+v", hostPlayer)
	}

	writeWSMessage(t, guestConn, wsMessage{
		Type:    messageTypeJoinRoom,
		Payload: joinRoomPayload{RoomCode: roomModel.Code},
	})

	guestSeesSelf := decodeRoomPlayer(t, readWSMessage(t, guestConn).Payload)
	guestSeesHost := decodeRoomPlayer(t, readWSMessage(t, guestConn).Payload)
	if guestSeesSelf.ID != guest.ID && guestSeesHost.ID != guest.ID {
		t.Fatalf("expected guest to receive their own join payload")
	}
	if guestSeesSelf.ID != host.ID && guestSeesHost.ID != host.ID {
		t.Fatalf("expected guest to receive existing host payload")
	}

	hostSeesGuest := readWSMessage(t, hostConn)
	if hostSeesGuest.Type != messageTypePlayerJoined {
		t.Fatalf("expected host to receive guest join, got %q", hostSeesGuest.Type)
	}
	if decodeRoomPlayer(t, hostSeesGuest.Payload).ID != guest.ID {
		t.Fatalf("expected host to receive guest player update")
	}

	writeWSMessage(t, guestConn, wsMessage{
		Type:    room.MessageTypeToggleReady,
		Payload: map[string]any{},
	})

	hostReadyUpdate := readWSMessage(t, hostConn)
	guestReadyUpdate := readWSMessage(t, guestConn)
	for _, update := range []wsMessage{hostReadyUpdate, guestReadyUpdate} {
		if update.Type != room.MessageTypePlayerUpdate {
			t.Fatalf("expected ready update type %q, got %q", room.MessageTypePlayerUpdate, update.Type)
		}
		player := decodeRoomPlayer(t, update.Payload)
		if player.ID != guest.ID || !player.IsReady {
			t.Fatalf("unexpected ready update payload: %+v", player)
		}
	}

	if err := guestConn.Close(); err != nil {
		t.Fatalf("failed to close guest websocket: %v", err)
	}

	hostLeftUpdate := readWSMessage(t, hostConn)
	if hostLeftUpdate.Type != messageTypePlayerLeft {
		t.Fatalf("expected player left message, got %q", hostLeftUpdate.Type)
	}
	left := decodePlayerLeft(t, hostLeftUpdate.Payload)
	if left.ID != guest.ID {
		t.Fatalf("expected guest ID %d in PLAYER_LEFT, got %d", guest.ID, left.ID)
	}
}

func TestWebSocketHandlerReconnectResetsReadyState(t *testing.T) {
	database := setupWebSocketTestDB(t)
	resetWebSocketStateForTesting()

	host := createWebSocketUser(t, database, "host-reset")
	guest := createWebSocketUser(t, database, "guest-reset")
	roomModel, err := room.Create(database, room.CreateParams{
		Code:       "ZX98QP",
		HostUserID: host.ID,
		MaxPlayers: 6,
	})
	if err != nil {
		t.Fatalf("failed to create room: %v", err)
	}

	server := httptest.NewServer(NewRouter())
	defer server.Close()

	hostConn := dialWebSocket(t, server.URL, sessionCookieForTest(t, "host-reset"))
	defer hostConn.Close()
	firstGuestConn := dialWebSocket(t, server.URL, sessionCookieForTest(t, "guest-reset"))

	writeWSMessage(t, hostConn, wsMessage{Type: messageTypeJoinRoom, Payload: joinRoomPayload{RoomCode: roomModel.Code}})
	_ = readWSMessage(t, hostConn)

	writeWSMessage(t, firstGuestConn, wsMessage{Type: messageTypeJoinRoom, Payload: joinRoomPayload{RoomCode: roomModel.Code}})
	_ = readWSMessage(t, firstGuestConn)
	_ = readWSMessage(t, firstGuestConn)
	_ = readWSMessage(t, hostConn)

	writeWSMessage(t, firstGuestConn, wsMessage{Type: room.MessageTypeToggleReady, Payload: map[string]any{}})
	_ = readWSMessage(t, firstGuestConn)
	_ = readWSMessage(t, hostConn)

	if err := firstGuestConn.Close(); err != nil {
		t.Fatalf("failed to close first guest connection: %v", err)
	}
	_ = readWSMessage(t, hostConn)

	secondGuestConn := dialWebSocket(t, server.URL, sessionCookieForTest(t, "guest-reset"))
	defer secondGuestConn.Close()

	writeWSMessage(t, secondGuestConn, wsMessage{Type: messageTypeJoinRoom, Payload: joinRoomPayload{RoomCode: roomModel.Code}})

	first := decodeRoomPlayer(t, readWSMessage(t, secondGuestConn).Payload)
	second := decodeRoomPlayer(t, readWSMessage(t, secondGuestConn).Payload)
	rejoined := first
	if rejoined.ID != guest.ID {
		rejoined = second
	}
	if rejoined.ID != guest.ID {
		t.Fatalf("expected rejoined guest payload, got %+v %+v", first, second)
	}
	if rejoined.IsReady {
		t.Fatal("expected ready state to reset on reconnect")
	}
}

func setupWebSocketTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	originalAppEnv, hadAppEnv := os.LookupEnv("APP_ENV")
	originalGoEnv, hadGoEnv := os.LookupEnv("GO_ENV")
	if err := os.Setenv("APP_ENV", "test"); err != nil {
		t.Fatalf("failed to set APP_ENV: %v", err)
	}
	if err := os.Setenv("GO_ENV", "test"); err != nil {
		t.Fatalf("failed to set GO_ENV: %v", err)
	}

	db.ResetForTesting()
	t.Cleanup(func() {
		_ = db.Close()
		resetWebSocketStateForTesting()
		if hadAppEnv {
			_ = os.Setenv("APP_ENV", originalAppEnv)
		} else {
			_ = os.Unsetenv("APP_ENV")
		}
		if hadGoEnv {
			_ = os.Setenv("GO_ENV", originalGoEnv)
		} else {
			_ = os.Unsetenv("GO_ENV")
		}
	})

	cfg := &db.Config{
		DatabasePath:    filepath.Join(t.TempDir(), "ws-test.db"),
		MaxOpenConns:    10,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Minute,
		LogLevel:        logger.Silent,
	}

	database, err := db.Connect(cfg)
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	return database
}

func createWebSocketUser(t *testing.T, database *gorm.DB, username string) *models.User {
	t.Helper()

	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	user := &models.User{
		Username:     username,
		PasswordHash: string(hash),
	}
	if err := database.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	return user
}

func sessionCookieForTest(t *testing.T, username string) *http.Cookie {
	t.Helper()

	rec := httptest.NewRecorder()
	if err := auth.SetSessionCookie(rec, username); err != nil {
		t.Fatalf("failed to set session cookie: %v", err)
	}

	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected session cookie")
	}

	return cookies[0]
}

func dialWebSocket(t *testing.T, serverURL string, cookie *http.Cookie) *websocket.Conn {
	t.Helper()

	wsURL := "ws" + serverURL[len("http"):] + "/ws"
	config, err := websocket.NewConfig(wsURL, "http://localhost:4200")
	if err != nil {
		t.Fatalf("failed to create websocket config: %v", err)
	}
	config.Header = http.Header{}
	config.Header.Set("Cookie", cookie.String())

	conn, err := websocket.DialConfig(config)
	if err != nil {
		t.Fatalf("failed to dial websocket: %v", err)
	}

	return conn
}

func writeWSMessage(t *testing.T, conn *websocket.Conn, message wsMessage) {
	t.Helper()

	if err := websocket.JSON.Send(conn, message); err != nil {
		t.Fatalf("failed to send websocket message: %v", err)
	}
}

func readWSMessage(t *testing.T, conn *websocket.Conn) wsMessage {
	t.Helper()

	if err := conn.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("failed to set websocket deadline: %v", err)
	}

	var message wsMessage
	if err := websocket.JSON.Receive(conn, &message); err != nil {
		t.Fatalf("failed to read websocket message: %v", err)
	}

	return message
}

func decodeRoomPlayer(t *testing.T, payload interface{}) room.Player {
	t.Helper()

	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map payload, got %T", payload)
	}

	return room.Player{
		ID:       uint(payloadMap["id"].(float64)),
		Username: payloadMap["username"].(string),
		IsHost:   payloadMap["isHost"].(bool),
		IsReady:  payloadMap["isReady"].(bool),
	}
}

func decodePlayerLeft(t *testing.T, payload interface{}) playerLeftPayload {
	t.Helper()

	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map payload, got %T", payload)
	}

	return playerLeftPayload{
		ID: uint(payloadMap["id"].(float64)),
	}
}
