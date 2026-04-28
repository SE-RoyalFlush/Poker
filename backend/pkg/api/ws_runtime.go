package api

import "github.com/SE-RoyalFlush/Poker/backend/pkg/socket"

type wsMessage = socket.Message

type wsHubAdapter struct{}

var globalWSHub wsHubAdapter

func resetWebSocketStateForTesting() {
	socket.ResetForTesting()
}

func (wsHubAdapter) activePlayerCount(roomCode string) int {
	return socket.Occupancy(roomCode)
}

func (wsHubAdapter) occupancy(roomCode string) int {
	return socket.Occupancy(roomCode)
}
