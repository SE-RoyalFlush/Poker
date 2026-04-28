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
	"github.com/SE-RoyalFlush/Poker/backend/pkg/protocol"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/room"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/bcrypt"
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
		Type:    protocol.ClientMsgJoinRoom,
		Payload: protocol.JoinRoomPayload{RoomCode: roomModel.Code},
	})

	hostSnapshot := readWSMessage(t, hostConn)
	if hostSnapshot.Type != protocol.ServerMsgRoomState {
		t.Fatalf("expected host snapshot message type %q, got %q", protocol.ServerMsgRoomState, hostSnapshot.Type)
	}
	hostPlayers := decodeRoomState(t, hostSnapshot.Payload)
	if len(hostPlayers) != 1 {
		t.Fatalf("expected snapshot with 1 player, got %d", len(hostPlayers))
	}
	if hostPlayers[0].Username != "host-user" || !hostPlayers[0].IsHost || hostPlayers[0].IsReady {
		t.Fatalf("unexpected host player payload: %+v", hostPlayers[0])
	}

	writeWSMessage(t, guestConn, wsMessage{
		Type:    protocol.ClientMsgJoinRoom,
		Payload: protocol.JoinRoomPayload{RoomCode: roomModel.Code},
	})

	guestSnapshot := readWSMessage(t, guestConn)
	if guestSnapshot.Type != protocol.ServerMsgRoomState {
		t.Fatalf("expected guest snapshot message type %q, got %q", protocol.ServerMsgRoomState, guestSnapshot.Type)
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
	if hostSeesGuest.Type != protocol.ServerMsgPlayerJoined {
		t.Fatalf("expected host to receive guest join, got %q", hostSeesGuest.Type)
	}
	if decodeRoomPlayer(t, hostSeesGuest.Payload).ID != guest.ID {
		t.Fatalf("expected host to receive guest player update")
	}

	writeWSMessage(t, guestConn, wsMessage{
		Type:    protocol.ClientMsgLeaveRoom,
		Payload: map[string]any{},
	})

	hostLeftUpdate := readWSMessage(t, hostConn)
	if hostLeftUpdate.Type != protocol.ServerMsgPlayerLeft {
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

	writeWSMessage(t, hostConn, wsMessage{Type: protocol.ClientMsgJoinRoom, Payload: protocol.JoinRoomPayload{RoomCode: roomModel.Code}})
	_ = readWSMessage(t, hostConn)

	writeWSMessage(t, firstGuestConn, wsMessage{Type: protocol.ClientMsgJoinRoom, Payload: protocol.JoinRoomPayload{RoomCode: roomModel.Code}})
	_ = readWSMessage(t, firstGuestConn)
	_ = readWSMessage(t, hostConn)

	writeWSMessage(t, firstGuestConn, wsMessage{Type: protocol.ClientMsgToggleReady, Payload: map[string]any{}})
	_ = readWSMessage(t, firstGuestConn)
	_ = readWSMessage(t, hostConn)

	if err := firstGuestConn.Close(); err != nil {
		t.Fatalf("failed to close first guest connection: %v", err)
	}
	_ = readWSMessage(t, hostConn)

	secondGuestConn := dialWebSocket(t, server.URL, sessionCookieForTest(t, "guest-reset"))
	defer secondGuestConn.Close()

	writeWSMessage(t, secondGuestConn, wsMessage{Type: protocol.ClientMsgJoinRoom, Payload: protocol.JoinRoomPayload{RoomCode: roomModel.Code}})

	snapshot := readWSMessage(t, secondGuestConn)
	if snapshot.Type != protocol.ServerMsgRoomState {
		t.Fatalf("expected rejoin snapshot type %q, got %q", protocol.ServerMsgRoomState, snapshot.Type)
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

	writeWSMessage(t, roomOneHostConn, wsMessage{Type: protocol.ClientMsgJoinRoom, Payload: protocol.JoinRoomPayload{RoomCode: roomOne.Code}})
	writeWSMessage(t, roomTwoHostConn, wsMessage{Type: protocol.ClientMsgJoinRoom, Payload: protocol.JoinRoomPayload{RoomCode: roomTwo.Code}})
	_ = readWSMessage(t, roomOneHostConn)
	_ = readWSMessage(t, roomTwoHostConn)

	writeWSMessage(t, roomOneGuestConn, wsMessage{Type: protocol.ClientMsgJoinRoom, Payload: protocol.JoinRoomPayload{RoomCode: roomOne.Code}})
	guestSnapshot := readWSMessage(t, roomOneGuestConn)
	if guestSnapshot.Type != protocol.ServerMsgRoomState {
		t.Fatalf("expected guest snapshot type %q, got %q", protocol.ServerMsgRoomState, guestSnapshot.Type)
	}

	hostOneEvent := readWSMessage(t, roomOneHostConn)
	if hostOneEvent.Type != protocol.ServerMsgPlayerJoined {
		t.Fatalf("expected room one host to receive %q, got %q", protocol.ServerMsgPlayerJoined, hostOneEvent.Type)
	}
	if decodeRoomPlayer(t, hostOneEvent.Payload).ID != guestRoomOne.ID {
		t.Fatalf("expected room one broadcast for guest ID %d", guestRoomOne.ID)
	}

	expectNoWSMessage(t, roomTwoHostConn)
}

// TestWebSocketHandlerUnknownMessageTypeReturnsError verifies that the server
// responds with an ERROR envelope when an unrecognized message type is received.
func TestWebSocketHandlerUnknownMessageTypeReturnsError(t *testing.T) {
	database := setupWebSocketTestDB(t)
	resetWebSocketStateForTesting()

	createWebSocketUser(t, database, "unknown-type-user")

	server := httptest.NewServer(NewRouter())
	defer server.Close()

	conn := dialWebSocket(t, server.URL, sessionCookieForTest(t, "unknown-type-user"))
	defer conn.Close()

	writeWSMessage(t, conn, wsMessage{
		Type:    "NOT_A_REAL_TYPE",
		Payload: map[string]any{},
	})

	resp := readWSMessage(t, conn)
	if resp.Type != protocol.ServerMsgError {
		t.Fatalf("expected %q for unknown message type, got %q", protocol.ServerMsgError, resp.Type)
	}

	errPayload := decodeErrorPayload(t, resp.Payload)
	if errPayload.Code != "UNKNOWN_TYPE" {
		t.Fatalf("expected error code %q, got %q", "UNKNOWN_TYPE", errPayload.Code)
	}
}

// TestWebSocketHandlerEnvelopeParsing verifies that the server correctly parses
// a valid {type, payload} envelope and routes it to the expected handler.
func TestWebSocketHandlerEnvelopeParsing(t *testing.T) {
	database := setupWebSocketTestDB(t)
	resetWebSocketStateForTesting()

	host := createWebSocketUser(t, database, "envelope-host")
	roomModel, err := room.Create(database, room.CreateParams{
		Code:       "ENVLP1",
		HostUserID: host.ID,
		MaxPlayers: 6,
	})
	if err != nil {
		t.Fatalf("failed to create room: %v", err)
	}

	server := httptest.NewServer(NewRouter())
	defer server.Close()

	conn := dialWebSocket(t, server.URL, sessionCookieForTest(t, "envelope-host"))
	defer conn.Close()

	// A valid envelope should be parsed and produce a ROOM_STATE response.
	writeWSMessage(t, conn, wsMessage{
		Type:    protocol.ClientMsgJoinRoom,
		Payload: protocol.JoinRoomPayload{RoomCode: roomModel.Code},
	})

	resp := readWSMessage(t, conn)
	if resp.Type != protocol.ServerMsgRoomState {
		t.Fatalf("expected %q after valid JOIN_ROOM envelope, got %q", protocol.ServerMsgRoomState, resp.Type)
	}
}

// TestWebSocketHandlerBroadcastShape verifies that server-emitted events conform
// to the protocol schema: envelope {type, payload} with the correct payload shape.
func TestWebSocketHandlerBroadcastShape(t *testing.T) {
	database := setupWebSocketTestDB(t)
	resetWebSocketStateForTesting()

	host := createWebSocketUser(t, database, "shape-host")
	guest := createWebSocketUser(t, database, "shape-guest")
	roomModel, err := room.Create(database, room.CreateParams{
		Code:       "SHAPE1",
		HostUserID: host.ID,
		MaxPlayers: 6,
	})
	if err != nil {
		t.Fatalf("failed to create room: %v", err)
	}

	server := httptest.NewServer(NewRouter())
	defer server.Close()

	hostConn := dialWebSocket(t, server.URL, sessionCookieForTest(t, "shape-host"))
	defer hostConn.Close()
	guestConn := dialWebSocket(t, server.URL, sessionCookieForTest(t, "shape-guest"))
	defer guestConn.Close()

	writeWSMessage(t, hostConn, wsMessage{Type: protocol.ClientMsgJoinRoom, Payload: protocol.JoinRoomPayload{RoomCode: roomModel.Code}})
	roomStateMsg := readWSMessage(t, hostConn)
	if roomStateMsg.Type != protocol.ServerMsgRoomState {
		t.Fatalf("expected %q, got %q", protocol.ServerMsgRoomState, roomStateMsg.Type)
	}
	players := decodeRoomState(t, roomStateMsg.Payload)
	if len(players) != 1 || players[0].ID != host.ID {
		t.Fatalf("ROOM_STATE payload malformed: %+v", players)
	}

	writeWSMessage(t, guestConn, wsMessage{Type: protocol.ClientMsgJoinRoom, Payload: protocol.JoinRoomPayload{RoomCode: roomModel.Code}})
	_ = readWSMessage(t, guestConn) // guest's own ROOM_STATE

	playerJoinedMsg := readWSMessage(t, hostConn)
	if playerJoinedMsg.Type != protocol.ServerMsgPlayerJoined {
		t.Fatalf("expected %q, got %q", protocol.ServerMsgPlayerJoined, playerJoinedMsg.Type)
	}
	joined := decodeRoomPlayer(t, playerJoinedMsg.Payload)
	if joined.ID != guest.ID || joined.Username != "shape-guest" {
		t.Fatalf("PLAYER_JOINED payload malformed: %+v", joined)
	}

	writeWSMessage(t, guestConn, wsMessage{Type: protocol.ClientMsgToggleReady, Payload: map[string]any{}})
	playerUpdateMsg := readWSMessage(t, guestConn)
	if playerUpdateMsg.Type != protocol.ServerMsgPlayerUpdate {
		t.Fatalf("expected %q, got %q", protocol.ServerMsgPlayerUpdate, playerUpdateMsg.Type)
	}
	updated := decodeRoomPlayer(t, playerUpdateMsg.Payload)
	if updated.ID != guest.ID || !updated.IsReady {
		t.Fatalf("PLAYER_UPDATE payload malformed: %+v", updated)
	}
}

func TestWebSocketHandlerRespondsToPing(t *testing.T) {
	database := setupWebSocketTestDB(t)
	resetWebSocketStateForTesting()
	createWebSocketUser(t, database, "ping-user")

	server := httptest.NewServer(NewRouter())
	defer server.Close()

	conn := dialWebSocket(t, server.URL, sessionCookieForTest(t, "ping-user"))
	defer conn.Close()

	pongReceived := make(chan struct{})
	conn.SetPongHandler(func(string) error {
		close(pongReceived)
		return nil
	})

	if err := conn.WriteControl(websocket.PingMessage, []byte("heartbeat"), time.Now().Add(time.Second)); err != nil {
		t.Fatalf("failed to send websocket ping: %v", err)
	}

	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		_, _, _ = conn.ReadMessage()
	}()

	select {
	case <-pongReceived:
	case <-time.After(2 * time.Second):
		t.Fatal("expected websocket pong response")
	}

	_ = conn.Close()
	<-readDone
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
	header := http.Header{}
	header.Set("Cookie", cookie.String())
	header.Set("Origin", "http://localhost:4200")

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		t.Fatalf("failed to dial websocket: %v", err)
	}

	return conn
}

func writeWSMessage(t *testing.T, conn *websocket.Conn, message wsMessage) {
	t.Helper()

	if err := conn.WriteJSON(message); err != nil {
		t.Fatalf("failed to send websocket message: %v", err)
	}
}

func readWSMessage(t *testing.T, conn *websocket.Conn) wsMessage {
	t.Helper()

	if err := conn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("failed to set websocket read deadline: %v", err)
	}

	var message wsMessage
	if err := conn.ReadJSON(&message); err != nil {
		t.Fatalf("failed to read websocket message: %v", err)
	}

	return message
}

func expectNoWSMessage(t *testing.T, conn *websocket.Conn) {
	t.Helper()

	if err := conn.SetReadDeadline(time.Now().Add(500 * time.Millisecond)); err != nil {
		t.Fatalf("failed to set websocket read deadline: %v", err)
	}
	defer func() {
		_ = conn.SetReadDeadline(time.Time{})
	}()

	var message wsMessage
	if err := conn.ReadJSON(&message); err == nil {
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

func decodeRoomPlayer(t *testing.T, payload interface{}) protocol.Player {
	t.Helper()

	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map payload, got %T", payload)
	}

	return protocol.Player{
		ID:       uint(payloadMap["id"].(float64)),
		Username: payloadMap["username"].(string),
		IsHost:   payloadMap["isHost"].(bool),
		IsReady:  payloadMap["isReady"].(bool),
	}
}

func decodePlayerLeft(t *testing.T, payload interface{}) protocol.PlayerLeftPayload {
	t.Helper()

	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map payload, got %T", payload)
	}

	return protocol.PlayerLeftPayload{
		ID: uint(payloadMap["id"].(float64)),
	}
}

func decodeErrorPayload(t *testing.T, payload interface{}) protocol.ErrorPayload {
	t.Helper()

	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map payload for error, got %T", payload)
	}

	return protocol.ErrorPayload{
		Code:    payloadMap["code"].(string),
		Message: payloadMap["message"].(string),
	}
}

func decodeRoomState(t *testing.T, payload interface{}) []protocol.Player {
	t.Helper()

	payloadMap, ok := payload.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map payload, got %T", payload)
	}

	rawPlayers, ok := payloadMap["players"].([]interface{})
	if !ok {
		t.Fatalf("expected players array, got %T", payloadMap["players"])
	}

	players := make([]protocol.Player, 0, len(rawPlayers))
	for _, rawPlayer := range rawPlayers {
		playerMap, ok := rawPlayer.(map[string]interface{})
		if !ok {
			t.Fatalf("expected player map, got %T", rawPlayer)
		}
		players = append(players, protocol.Player{
			ID:       uint(playerMap["id"].(float64)),
			Username: playerMap["username"].(string),
			IsHost:   playerMap["isHost"].(bool),
			IsReady:  playerMap["isReady"].(bool),
		})
	}

	return players
}

func containsPlayer(players []protocol.Player, playerID uint) bool {
	_, ok := findPlayer(players, playerID)
	return ok
}

func findPlayer(players []protocol.Player, playerID uint) (protocol.Player, bool) {
	for _, player := range players {
		if player.ID == playerID {
			return player, true
		}
	}

	return protocol.Player{}, false
}
