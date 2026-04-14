package room

import (
	"errors"
	"sync"
)

var (
	ErrPlayerNotFound = errors.New("player not found")
)

const (
	MessageTypeToggleReady  = "TOGGLE_READY"
	MessageTypePlayerUpdate = "PLAYER_UPDATE"
)

// Player represents the public player state shared with room participants.
type Player struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	IsHost   bool   `json:"isHost"`
	IsReady  bool   `json:"isReady"`
}

// Event represents a websocket-style room event envelope.
type Event struct {
	Type    string `json:"type"`
	Payload Player `json:"payload"`
}

// Lobby stores the in-memory player state for a room.
type Lobby struct {
	mu      sync.RWMutex
	players map[uint]Player
}

// NewLobby creates an empty in-memory room lobby state.
func NewLobby() *Lobby {
	return &Lobby{
		players: make(map[uint]Player),
	}
}

// JoinPlayer adds a player to the lobby or resets their ready state on rejoin.
func (l *Lobby) JoinPlayer(player Player) Player {
	l.mu.Lock()
	defer l.mu.Unlock()

	player.IsReady = false
	l.players[player.ID] = player

	return player
}

// RemovePlayer removes a player from the lobby state.
func (l *Lobby) RemovePlayer(playerID uint) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.players, playerID)
}

// ToggleReady flips a player's ready state and returns the event to broadcast.
func (l *Lobby) ToggleReady(playerID uint) (Event, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	player, ok := l.players[playerID]
	if !ok {
		return Event{}, ErrPlayerNotFound
	}

	player.IsReady = !player.IsReady
	l.players[playerID] = player

	return Event{
		Type:    MessageTypePlayerUpdate,
		Payload: player,
	}, nil
}

// HandleMessage applies supported room messages and returns any event to broadcast.
func (l *Lobby) HandleMessage(messageType string, playerID uint) (Event, error) {
	switch messageType {
	case MessageTypeToggleReady:
		return l.ToggleReady(playerID)
	default:
		return Event{}, nil
	}
}

// Player returns the current player state.
func (l *Lobby) Player(playerID uint) (Player, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	player, ok := l.players[playerID]
	return player, ok
}

// Players returns a snapshot of the current room roster.
func (l *Lobby) Players() []Player {
	l.mu.RLock()
	defer l.mu.RUnlock()

	players := make([]Player, 0, len(l.players))
	for _, player := range l.players {
		players = append(players, player)
	}

	return players
}

// AllReady reports whether every tracked player is marked ready.
func (l *Lobby) AllReady() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if len(l.players) == 0 {
		return false
	}

	for _, player := range l.players {
		if !player.IsReady {
			return false
		}
	}

	return true
}
