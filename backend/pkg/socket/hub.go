package socket

import (
	"errors"
	"log"
	"sync"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/protocol"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/room"
	"gorm.io/gorm"
)

var errMissingRoomCode = errors.New("missing room code")

type Message = protocol.Envelope[any]

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

func Occupancy(roomCode string) int {
	return defaultHub.Occupancy(roomCode)
}

func RoomOccupancy(roomCode string) int {
	return defaultHub.Occupancy(roomCode)
}

func (h *Hub) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.rooms = make(map[string]*roomState)
}

func (h *Hub) Occupancy(roomCode string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	state, ok := h.rooms[roomCode]
	if !ok {
		return 0
	}

	uniquePlayers := make(map[uint]struct{}, len(state.clients))
	for client := range state.clients {
		if client != nil && client.user != nil {
			uniquePlayers[client.user.ID] = struct{}{}
		}
	}

	return len(uniquePlayers)
}

func toProtocolPlayer(p room.Player) protocol.Player {
	return protocol.Player{
		ID:       p.ID,
		Username: p.Username,
		IsHost:   p.IsHost,
		IsReady:  p.IsReady,
	}
}

func toProtocolPlayers(players []room.Player) []protocol.Player {
	result := make([]protocol.Player, len(players))
	for i, p := range players {
		result[i] = toProtocolPlayer(p)
	}
	return result
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

	client.setRoomCode(roomModel.Code)
	state.clients[client] = true

	player := state.lobby.JoinPlayer(room.Player{
		ID:       client.user.ID,
		Username: client.user.Username,
		IsHost:   roomModel.HostUserID == client.user.ID,
	})

	broadcast := Message{
		Type:    protocol.ServerMsgPlayerJoined,
		Payload: toProtocolPlayer(player),
	}

	recipients := make([]*Client, 0, len(state.clients))
	for member := range state.clients {
		if member == client {
			continue
		}
		recipients = append(recipients, member)
	}

	snapshot := Message{
		Type: protocol.ServerMsgRoomState,
		Payload: protocol.RoomStatePayload{
			RoomCode: roomModel.Code,
			Players:  toProtocolPlayers(state.lobby.Players()),
		},
	}

	h.mu.Unlock()

	for _, member := range recipients {
		member.trySend(broadcast)
	}

	return []Message{snapshot}, nil
}

func (h *Hub) Leave(client *Client) {
	if client == nil {
		return
	}

	roomCode := client.roomCodeValue()
	if roomCode == "" {
		return
	}

	h.mu.Lock()

	state, ok := h.rooms[roomCode]
	if !ok {
		client.setRoomCode("")
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
			Type:    protocol.ServerMsgPlayerLeft,
			Payload: protocol.PlayerLeftPayload{ID: client.user.ID},
		}
		recipients = make([]*Client, 0, len(state.clients))
		for member := range state.clients {
			recipients = append(recipients, member)
		}
	}

	shouldCleanupRoom := len(state.clients) == 0
	if shouldCleanupRoom {
		delete(h.rooms, roomCode)
	}

	client.setRoomCode("")
	h.mu.Unlock()

	if shouldCleanupRoom {
		h.markRoomInactive(roomCode)
	}

	if !hasActiveConnection {
		for _, member := range recipients {
			member.trySend(leftMessage)
		}
	}
}

func (h *Hub) markRoomInactive(roomCode string) {
	if err := db.WithDB(func(database *gorm.DB) error {
		return room.MarkInactive(database, roomCode)
	}); err != nil {
		log.Printf("failed to mark room %s inactive: %v", roomCode, err)
	}
}

func (h *Hub) HandleRoomMessage(client *Client, messageType string) error {
	if client == nil {
		return nil
	}

	roomCode := client.roomCodeValue()
	if roomCode == "" {
		return nil
	}

	h.mu.RLock()
	state, ok := h.rooms[roomCode]
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
		Payload: toProtocolPlayer(event.Payload),
	}

	h.mu.RLock()
	recipients := make([]*Client, 0, len(state.clients))
	for member := range state.clients {
		recipients = append(recipients, member)
	}
	h.mu.RUnlock()

	for _, member := range recipients {
		member.trySend(message)
	}

	return nil
}
