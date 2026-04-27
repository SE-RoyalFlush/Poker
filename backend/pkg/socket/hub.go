package socket

import (
	"errors"
	"sync"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/room"
)

const (
	MessageTypeJoinRoom     = "JOIN_ROOM"
	MessageTypePlayerJoined = "PLAYER_JOINED"
	MessageTypePlayerLeft   = "PLAYER_LEFT"
)

var errMissingRoomCode = errors.New("missing room code")

type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type JoinRoomPayload struct {
	RoomCode string `json:"roomCode"`
}

type PlayerLeftPayload struct {
	ID uint `json:"id"`
}

type roomState struct {
	lobby   *room.Lobby
	clients map[*Client]bool
}

type Hub struct {
	mu    sync.RWMutex
	rooms map[string]*roomState
}

func NewHub() *Hub {
	return &Hub{
		rooms: make(map[string]*roomState),
	}
}

var defaultHub = NewHub()

func ResetForTesting() {
	defaultHub.Reset()
}

func (h *Hub) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.rooms = make(map[string]*roomState)
}

func (h *Hub) Join(roomModel *models.Room, client *Client) ([]Message, error) {
	if roomModel == nil {
		return nil, errMissingRoomCode
	}

	h.mu.Lock()

	state, ok := h.rooms[roomModel.Code]
	if !ok {
		state = &roomState{
			lobby:   room.NewLobby(),
			clients: make(map[*Client]bool),
		}
		h.rooms[roomModel.Code] = state
	}

	client.roomCode = roomModel.Code
	state.clients[client] = true

	player := state.lobby.JoinPlayer(room.Player{
		ID:       client.user.ID,
		Username: client.user.Username,
		IsHost:   roomModel.HostUserID == client.user.ID,
	})

	existingPlayers := make([]room.Player, 0, len(state.clients))
	for existingClient := range state.clients {
		if existingClient == client {
			continue
		}
		existingPlayer, ok := state.lobby.Player(existingClient.user.ID)
		if ok {
			existingPlayers = append(existingPlayers, existingPlayer)
		}
	}

	broadcast := Message{
		Type:    MessageTypePlayerJoined,
		Payload: player,
	}

	recipients := make([]*Client, 0, len(state.clients))
	for member := range state.clients {
		recipients = append(recipients, member)
	}

	h.mu.Unlock()

	for _, member := range recipients {
		member.send <- broadcast
	}

	initialMessages := make([]Message, 0, len(existingPlayers))
	for _, existingPlayer := range existingPlayers {
		initialMessages = append(initialMessages, Message{
			Type:    MessageTypePlayerJoined,
			Payload: existingPlayer,
		})
	}

	return initialMessages, nil
}

func (h *Hub) Leave(client *Client) {
	if client == nil || client.roomCode == "" {
		return
	}

	h.mu.Lock()

	state, ok := h.rooms[client.roomCode]
	if !ok {
		client.roomCode = ""
		h.mu.Unlock()
		return
	}

	delete(state.clients, client)

	hasActiveConnection := false
	for member := range state.clients {
		if member != nil && member.user.ID == client.user.ID {
			hasActiveConnection = true
			break
		}
	}

	var recipients []*Client
	var leftMessage Message
	if !hasActiveConnection {
		state.lobby.RemovePlayer(client.user.ID)

		leftMessage = Message{
			Type:    MessageTypePlayerLeft,
			Payload: PlayerLeftPayload{ID: client.user.ID},
		}
		recipients = make([]*Client, 0, len(state.clients))
		for member := range state.clients {
			recipients = append(recipients, member)
		}
	}

	if len(state.clients) == 0 {
		delete(h.rooms, client.roomCode)
	}

	client.roomCode = ""
	h.mu.Unlock()

	if !hasActiveConnection {
		for _, member := range recipients {
			member.send <- leftMessage
		}
	}
}

func (h *Hub) HandleRoomMessage(client *Client, messageType string) error {
	if client == nil || client.roomCode == "" {
		return nil
	}

	h.mu.RLock()
	state, ok := h.rooms[client.roomCode]
	h.mu.RUnlock()
	if !ok {
		return nil
	}

	event, err := state.lobby.HandleMessage(messageType, client.user.ID)
	if err != nil {
		return err
	}
	if event.Type == "" {
		return nil
	}

	message := Message{
		Type:    event.Type,
		Payload: event.Payload,
	}

	h.mu.RLock()
	recipients := make([]*Client, 0, len(state.clients))
	for member := range state.clients {
		recipients = append(recipients, member)
	}
	h.mu.RUnlock()

	for _, member := range recipients {
		member.send <- message
	}

	return nil
}
