package api

import (
	"errors"
	"io"
	"net"
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

	hostSnapshot := readWSMessage(t, hostConn)
	if hostSnapshot.Type != messageTypeRoomState {
		t.Fatalf("expected host snapshot message type %q, got %q", messageTypeRoomState, hostSnapshot.Type)
	}
	hostPlayers := decodeRoomState(t, hostSnapshot.Payload)
	if len(hostPlayers) != 1 {
		t.Fatalf("expected snapshot with 1 player, got %d", len(hostPlayers))
	}
	if hostPlayers[0].Username != "host-user" || !hostPlayers[0].IsHost || hostPlayers[0].IsReady {
		t.Fatalf("unexpected host player payload: %+v", hostPlayers[0])
	}

	writeWSMessage(t, guestConn, wsMessage{
		Type:    messageTypeJoinRoom,
		Payload: joinRoomPayload{RoomCode: roomModel.Code},
	})

	guestSnapshot := readWSMessage(t, guestConn)
	if guestSnapshot.Type != messageTypeRoomState {
		t.Fatalf("expected guest snapshot message type %q, got %q", messageTypeRoomState, guestSnapshot.Type)
	}
	guestPlayers := decodeRoomState(t, guestSnapshot.Payload)
	if len(guestPlayers) != 2 {
		t.Fatalf("expected snapshot with 2 players, got %d", len(guestPlayers))
	}
	if !containsPlayer(guestPlayers, guest.ID) {
		t.Fatalf("expected guest snapshot to include guest ID %d", guest.ID)
	}
	if !containsPlayer(guestPlayers, host.ID) {
		t.Fatalf("expected guest snapshot to include host ID %d", host.ID)
	}

	hostSeesGuest := readWSMessage(t, hostConn)
	if hostSeesGuest.Type != messageTypePlayerJoined {
		t.Fatalf("expected host to receive guest join, got %q", hostSeesGuest.Type)
	}
	if decodeRoomPlayer(t, hostSeesGuest.Payload).ID != guest.ID {
		t.Fatalf("expected host to receive guest player update")
	}

	writeWSMessage(t, guestConn, wsMessage{
		Type:    messageTypeLeaveRoom,
		Payload: map[string]any{},
	})

	hostLeftUpdate := readWSMessage(t, hostConn)
	if hostLeftUpdate.Type != messageTypePlayerLeft {
		t.Fatalf("expected player left message, got %q", hostLeftUpdate.Type)
	}
	left := decodePlayerLeft(t, hostLeftUpdate.Payload)
	if left.ID != guest.ID {
		t.Fatalf("expected guest ID %d in PLAYER_LEFT, got %d", guest.ID, left.ID)
	}

	expectNoWSMessage(t, guestConn)
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

	snapshot := readWSMessage(t, secondGuestConn)
	if snapshot.Type != messageTypeRoomState {
		t.Fatalf("expected rejoin snapshot type %q, got %q", messageTypeRoomState, snapshot.Type)
	}
	players := decodeRoomState(t, snapshot.Payload)
	rejoined, ok := findPlayer(players, guest.ID)
	if !ok {
		t.Fatalf("expected rejoined guest payload in snapshot, got %+v", players)
	}
	if rejoined.IsReady {
		t.Fatal("expected ready state to reset on reconnect")
	}
}

func TestWebSocketHandlerIsolatesRoomsAndBroadcastsWithinRoomOnly(t *testing.T) {
	database := setupWebSocketTestDB(t)
	resetWebSocketStateForTesting()

	hostRoomOne := createWebSocketUser(t, database, "host-room-one")
	hostRoomTwo := createWebSocketUser(t, database, "host-room-two")
	guestRoomOne := createWebSocketUser(t, database, "guest-room-one")

	roomOne, err := room.Create(database, room.CreateParams{
		Code:       "ROOM11",
		HostUserID: hostRoomOne.ID,
		MaxPlayers: 6,
	})
	if err != nil {
		t.Fatalf("failed to create room one: %v", err)
	}
	roomTwo, err := room.Create(database, room.CreateParams{
		Code:       "ROOM22",
		HostUserID: hostRoomTwo.ID,
		MaxPlayers: 6,
	})
	if err != nil {
		t.Fatalf("failed to create room two: %v", err)
	}

	server := httptest.NewServer(NewRouter())
	defer server.Close()

	roomOneHostConn := dialWebSocket(t, server.URL, sessionCookieForTest(t, "host-room-one"))
	defer roomOneHostConn.Close()
	roomTwoHostConn := dialWebSocket(t, server.URL, sessionCookieForTest(t, "host-room-two"))
	defer roomTwoHostConn.Close()
	roomOneGuestConn := dialWebSocket(t, server.URL, sessionCookieForTest(t, "guest-room-one"))
	defer roomOneGuestConn.Close()

	writeWSMessage(t, roomOneHostConn, wsMessage{Type: messageTypeJoinRoom, Payload: joinRoomPayload{RoomCode: roomOne.Code}})
	writeWSMessage(t, roomTwoHostConn, wsMessage{Type: messageTypeJoinRoom, Payload: joinRoomPayload{RoomCode: roomTwo.Code}})
	_ = readWSMessage(t, roomOneHostConn)
	_ = readWSMessage(t, roomTwoHostConn)

	writeWSMessage(t, roomOneGuestConn, wsMessage{Type: messageTypeJoinRoom, Payload: joinRoomPayload{RoomCode: roomOne.Code}})
	guestSnapshot := readWSMessage(t, roomOneGuestConn)
	if guestSnapshot.Type != messageTypeRoomState {
		t.Fatalf("expected guest snapshot type %q, got %q", messageTypeRoomState, guestSnapshot.Type)
	}

	hostOneEvent := readWSMessage(t, roomOneHostConn)
	if hostOneEvent.Type != messageTypePlayerJoined {
		t.Fatalf("expected room one host to receive %q, got %q", messageTypePlayerJoined, hostOneEvent.Type)
	}
	if decodeRoomPlayer(t, hostOneEvent.Payload).ID != guestRoomOne.ID {
		t.Fatalf("expected room one broadcast for guest ID %d", guestRoomOne.ID)
	}

	expectNoWSMessage(t, roomTwoHostConn)
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

func expectNoWSMessage(t *testing.T, conn *websocket.Conn) {
	t.Helper()

	if err := conn.SetDeadline(time.Now().Add(500 * time.Millisecond)); err != nil {
		t.Fatalf("failed to set websocket deadline: %v", err)
	}
	defer func() {
		_ = conn.SetDeadline(time.Time{})
	}()

	var message wsMessage
	if err := websocket.JSON.Receive(conn, &message); err == nil {
		t.Fatalf("expected no websocket message, got %+v", message)
	} else if !isTimeoutError(err) {
		t.Fatalf("expected timeout while waiting for no websocket message, got %v", err)
	}
}

func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, os.ErrDeadlineExceeded) {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	if errors.Is(err, io.EOF) {
		return false
	}

	return false
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

func decodeRoomState(t *testing.T, payload interface{}) []room.Player {
	t.Helper()

	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map payload, got %T", payload)
	}

	rawPlayers, ok := payloadMap["players"].([]interface{})
	if !ok {
		t.Fatalf("expected players array, got %T", payloadMap["players"])
	}

	players := make([]room.Player, 0, len(rawPlayers))
	for _, rawPlayer := range rawPlayers {
		playerMap, ok := rawPlayer.(map[string]interface{})
		if !ok {
			t.Fatalf("expected player map, got %T", rawPlayer)
		}
		players = append(players, room.Player{
			ID:       uint(playerMap["id"].(float64)),
			Username: playerMap["username"].(string),
			IsHost:   playerMap["isHost"].(bool),
			IsReady:  playerMap["isReady"].(bool),
		})
	}

	return players
}

func containsPlayer(players []room.Player, playerID uint) bool {
	_, ok := findPlayer(players, playerID)
	return ok
}

func findPlayer(players []room.Player, playerID uint) (room.Player, bool) {
	for _, player := range players {
		if player.ID == playerID {
			return player, true
		}
	}

	return room.Player{}, false
}
