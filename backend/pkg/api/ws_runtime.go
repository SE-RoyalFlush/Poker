package api

import (
	"errors"
	"sync"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/room"
)

const (
	messageTypeJoinRoom     = "JOIN_ROOM"
	messageTypePlayerJoined = "PLAYER_JOINED"
	messageTypePlayerLeft   = "PLAYER_LEFT"
)

var (
	errMissingRoomCode = errors.New("missing room code")
)

type wsMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type joinRoomPayload struct {
	RoomCode string `json:"roomCode"`
}

type playerLeftPayload struct {
	ID uint `json:"id"`
}

type wsClient struct {
	user     *models.User
	send     chan wsMessage
	roomCode string
}

type wsRoom struct {
	lobby   *room.Lobby
	clients map[*wsClient]struct{}
}

type wsHub struct {
	mu    sync.RWMutex
	rooms map[string]*wsRoom
}

func newWSHub() *wsHub {
	return &wsHub{
		rooms: make(map[string]*wsRoom),
	}
}

var globalWSHub = newWSHub()

func resetWebSocketStateForTesting() {
	globalWSHub = newWSHub()
}

func (h *wsHub) join(roomModel *models.Room, client *wsClient) ([]wsMessage, error) {
	if roomModel == nil {
		return nil, errMissingRoomCode
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	wsRoomState, ok := h.rooms[roomModel.Code]
	if !ok {
		wsRoomState = &wsRoom{
			lobby:   room.NewLobby(),
			clients: make(map[*wsClient]struct{}),
		}
		h.rooms[roomModel.Code] = wsRoomState
	}

	client.roomCode = roomModel.Code
	wsRoomState.clients[client] = struct{}{}

	player := wsRoomState.lobby.JoinPlayer(room.Player{
		ID:       client.user.ID,
		Username: client.user.Username,
		IsHost:   roomModel.HostUserID == client.user.ID,
	})

	existingPlayers := make([]room.Player, 0, len(wsRoomState.clients))
	for existingClient := range wsRoomState.clients {
		if existingClient == client {
			continue
		}
		existingPlayer, ok := wsRoomState.lobby.Player(existingClient.user.ID)
		if ok {
			existingPlayers = append(existingPlayers, existingPlayer)
		}
	}

	broadcast := wsMessage{
		Type:    messageTypePlayerJoined,
		Payload: player,
	}

	for member := range wsRoomState.clients {
		member.send <- broadcast
	}

	initialMessages := make([]wsMessage, 0, len(existingPlayers))
	for _, existingPlayer := range existingPlayers {
		initialMessages = append(initialMessages, wsMessage{
			Type:    messageTypePlayerJoined,
			Payload: existingPlayer,
		})
	}

	return initialMessages, nil
}

func (h *wsHub) leave(client *wsClient) {
	if client == nil || client.roomCode == "" {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	wsRoomState, ok := h.rooms[client.roomCode]
	if !ok {
		client.roomCode = ""
		return
	}

	delete(wsRoomState.clients, client)
	wsRoomState.lobby.RemovePlayer(client.user.ID)

	leftMessage := wsMessage{
		Type:    messageTypePlayerLeft,
		Payload: playerLeftPayload{ID: client.user.ID},
	}
	for member := range wsRoomState.clients {
		member.send <- leftMessage
	}

	if len(wsRoomState.clients) == 0 {
		delete(h.rooms, client.roomCode)
	}

	client.roomCode = ""
}

func (h *wsHub) handleRoomMessage(client *wsClient, messageType string) error {
	if client == nil || client.roomCode == "" {
		return nil
	}

	h.mu.RLock()
	wsRoomState, ok := h.rooms[client.roomCode]
	h.mu.RUnlock()
	if !ok {
		return nil
	}

	event, err := wsRoomState.lobby.HandleMessage(messageType, client.user.ID)
	if err != nil {
		return err
	}
	if event.Type == "" {
		return nil
	}

	message := wsMessage{
		Type:    event.Type,
		Payload: event.Payload,
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for member := range wsRoomState.clients {
		member.send <- message
	}

	return nil
}
