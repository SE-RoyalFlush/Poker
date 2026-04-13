package api

import (
	"encoding/json"
	"net/http"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/auth"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/room"
	"golang.org/x/net/websocket"
)

type incomingWSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
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

	websocket.Handler(func(conn *websocket.Conn) {
		handleWebSocketConnection(conn, user)
	}).ServeHTTP(w, r)
}

func handleWebSocketConnection(conn *websocket.Conn, user *models.User) {
	client := &wsClient{
		user: user,
		send: make(chan wsMessage, 16),
	}

	defer func() {
		globalWSHub.leave(client)
		close(client.send)
		_ = conn.Close()
	}()

	go writeWebSocketMessages(conn, client.send)

	for {
		var message incomingWSMessage
		if err := websocket.JSON.Receive(conn, &message); err != nil {
			return
		}

		switch message.Type {
		case messageTypeJoinRoom:
			if err := handleJoinRoomMessage(client, message.Payload); err != nil {
				continue
			}
		case room.MessageTypeToggleReady:
			if err := globalWSHub.handleRoomMessage(client, message.Type); err != nil {
				continue
			}
		}
	}
}

func writeWebSocketMessages(conn *websocket.Conn, send <-chan wsMessage) {
	for message := range send {
		if err := websocket.JSON.Send(conn, message); err != nil {
			return
		}
	}
}

func handleJoinRoomMessage(client *wsClient, payload json.RawMessage) error {
	var joinPayload joinRoomPayload
	if err := json.Unmarshal(payload, &joinPayload); err != nil {
		return err
	}

	roomModel, err := loadRoomByCode(joinPayload.RoomCode)
	if err != nil {
		return err
	}

	initialMessages, err := globalWSHub.join(roomModel, client)
	if err != nil {
		return err
	}

	for _, message := range initialMessages {
		client.send <- message
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
