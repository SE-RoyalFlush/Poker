package socket

import (
	"errors"
	"log"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/game"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/protocol"
	"gorm.io/gorm"
)

var errHandNotComplete = errors.New("hand is not complete")

// CompleteFold resolves a hand when a player folds and another room participant
// remains. Persistence errors are logged by CompleteHand and do not stop the
// GAME_OVER broadcast.
func (h *Hub) CompleteFold(database *gorm.DB, client *Client) {
	payload, roomCode, err := h.foldResult(client)
	if err != nil {
		return
	}

	h.CompleteHand(database, roomCode, payload)
}

// CompleteHand persists the hand result and broadcasts GAME_OVER to room clients.
// Persistence errors are logged and do not interrupt gameplay.
func (h *Hub) CompleteHand(database *gorm.DB, roomCode string, payload protocol.GameOverPayload) {
	result := models.GameResult{
		WinnerID: payload.WinnerID,
		PotSize:  payload.Pot,
		GameType: models.DefaultGameType,
	}
	if err := game.SaveGameResult(database, result); err != nil {
		log.Printf("failed to persist game result for room %s: %v", roomCode, err)
	}

	h.BroadcastToRoom(roomCode, Message{
		Type:    protocol.ServerMsgGameOver,
		Payload: payload,
	})
}

func (h *Hub) foldResult(client *Client) (protocol.GameOverPayload, string, error) {
	if client == nil || client.user == nil {
		return protocol.GameOverPayload{}, "", errHandNotComplete
	}

	roomCode := client.roomCodeValue()
	if roomCode == "" {
		return protocol.GameOverPayload{}, "", errHandNotComplete
	}

	h.mu.RLock()
	state, ok := h.rooms[roomCode]
	if !ok {
		h.mu.RUnlock()
		return protocol.GameOverPayload{}, "", errHandNotComplete
	}
	players := toProtocolPlayers(state.lobby.Players())
	h.mu.RUnlock()

	if len(players) < 2 {
		return protocol.GameOverPayload{}, "", errHandNotComplete
	}

	var winnerID uint
	for _, player := range players {
		if player.ID != client.user.ID {
			winnerID = player.ID
			break
		}
	}
	if winnerID == 0 {
		return protocol.GameOverPayload{}, "", errHandNotComplete
	}

	return protocol.GameOverPayload{
		WinnerID: winnerID,
		Pot:      0,
		Seats:    players,
	}, roomCode, nil
}

// BroadcastToRoom sends a message to all currently connected clients in a room.
func (h *Hub) BroadcastToRoom(roomCode string, message Message) {
	h.mu.RLock()
	state, ok := h.rooms[roomCode]
	if !ok {
		h.mu.RUnlock()
		return
	}

	recipients := make([]*Client, 0, len(state.clients))
	for member := range state.clients {
		recipients = append(recipients, member)
	}
	h.mu.RUnlock()

	for _, member := range recipients {
		member.trySend(message)
	}
}
