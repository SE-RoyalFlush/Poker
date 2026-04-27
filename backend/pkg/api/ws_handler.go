package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/auth"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/protocol"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/room"
	"golang.org/x/net/websocket"
)

type incomingWSMessage = protocol.Envelope[json.RawMessage]

// allowedOrigins returns the list of origins permitted to open a WebSocket
// connection. The list is read from the ALLOWED_ORIGINS environment variable
// (comma-separated). When the variable is absent, it defaults to the local
// Angular development server.
func allowedOrigins() []string {
	raw := os.Getenv("ALLOWED_ORIGINS")
	if raw == "" {
		return []string{"http://localhost:4200"}
	}
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}

// isOriginAllowed reports whether the parsed WebSocket Origin is in the
// configured allowlist.
func isOriginAllowed(origin *url.URL) bool {
	if origin == nil {
		return false
	}
	originStr := origin.Scheme + "://" + origin.Host
	for _, allowed := range allowedOrigins() {
		if strings.TrimRight(allowed, "/") == originStr {
			return true
		}
	}
	return false
}

// WebSocketHandler upgrades the authenticated request and manages lobby events.
func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	user, err := auth.AuthenticatedUserFromRequest(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		sendError(w, "Unauthorized", "Authentication required", http.StatusUnauthorized)
		return
	}
	if _, ok := w.(http.Hijacker); !ok {
		w.Header().Set("Content-Type", "application/json")
		sendError(w, "Bad Request", "WebSocket upgrade is not supported", http.StatusBadRequest)
		return
	}

	wsServer := websocket.Server{
		Handler: func(conn *websocket.Conn) {
			handleWebSocketConnection(conn, user)
		},
		Handshake: func(cfg *websocket.Config, req *http.Request) error {
			origin, err := websocket.Origin(cfg, req)
			if err != nil {
				return fmt.Errorf("forbidden: invalid origin: %w", err)
			}
			if !isOriginAllowed(origin) {
				return fmt.Errorf("forbidden: origin %q is not allowed", origin)
			}
			return nil
		},
	}
	wsServer.ServeHTTP(w, r)
}

func handleWebSocketConnection(conn *websocket.Conn, user *models.User) {
	client := &wsClient{
		user: user,
		send: make(chan wsMessage, 16),
		done: make(chan struct{}),
	}

	defer func() {
		// Signal the writer goroutine to stop before leaving the hub, so
		// no further broadcasts can race against the connection teardown.
		close(client.done)
		globalWSHub.leave(client)
		_ = conn.Close()
	}()

	go writeWebSocketMessages(conn, client)

	for {
		var message incomingWSMessage
		if err := websocket.JSON.Receive(conn, &message); err != nil {
			return
		}

		switch message.Type {
		case protocol.ClientMsgJoinRoom:
			if err := handleJoinRoomMessage(client, message.Payload); err != nil {
				sendWSError(client, "JOIN_FAILED", err.Error())
			}
		case protocol.ClientMsgLeaveRoom:
			globalWSHub.leave(client)
		case protocol.ClientMsgToggleReady:
			if err := globalWSHub.handleRoomMessage(client, message.Type); err != nil {
				sendWSError(client, "ROOM_ACTION_FAILED", err.Error())
			}
		default:
			sendWSError(client, "UNKNOWN_TYPE", fmt.Sprintf("unrecognized message type: %q", message.Type))
		}
	}
}

func writeWebSocketMessages(conn *websocket.Conn, client *wsClient) {
	for {
		select {
		case <-client.done:
			return
		case message, ok := <-client.send:
			if !ok {
				return
			}
			if err := websocket.JSON.Send(conn, message); err != nil {
				return
			}
		}
	}
}

func handleJoinRoomMessage(client *wsClient, payload json.RawMessage) error {
	var joinPayload protocol.JoinRoomPayload
	if err := json.Unmarshal(payload, &joinPayload); err != nil {
		return err
	}

	roomModel, err := loadRoomByCode(joinPayload.RoomCode)
	if err != nil {
		return err
	}

	client.mu.Lock()
	currentRoom := client.roomCode
	client.mu.Unlock()

	if currentRoom != "" && currentRoom != roomModel.Code {
		globalWSHub.leave(client)
	}

	initialMessages, err := globalWSHub.join(roomModel, client)
	if err != nil {
		return err
	}

	for _, message := range initialMessages {
		select {
		case client.send <- message:
		case <-client.done:
			return nil
		}
	}

	return nil
}

func loadRoomByCode(code string) (*models.Room, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	return room.FindByCode(database, code)
}

func sendWSError(client *wsClient, code, msg string) {
	message := wsMessage{
		Type: protocol.ServerMsgError,
		Payload: protocol.ErrorPayload{
			Code:    code,
			Message: msg,
		},
	}

	select {
	case client.send <- message:
	case <-client.done:
	}
}
