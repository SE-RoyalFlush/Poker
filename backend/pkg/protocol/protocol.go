// Package protocol defines the WebSocket message contract shared between the
// game server and its clients.
//
// # Wire format
//
// Every message – sent by the client or broadcast by the server – uses a
// single envelope:
//
//	{ "type": "<EVENT_TYPE>", "payload": { ... } }
//
// # Client → Server events
//
//	JOIN_ROOM     { roomCode: string }
//	LEAVE_ROOM    {}
//	TOGGLE_READY  {}
//	FOLD          {}
//
// # Server → Client events
//
//	ROOM_STATE    { roomCode: string, players: Player[] }   (sent on join)
//	PLAYER_JOINED { id, username, isHost, isReady }         (broadcast to others)
//	PLAYER_LEFT   { id }                                    (broadcast to others)
//	PLAYER_UPDATE { id, username, isHost, isReady }         (broadcast to all)
//	GAME_OVER     { winnerId: number, pot: number, seats: Player[] }
//	ERROR         { code: string, message: string }
package protocol

// Client-to-server message type constants.
const (
	ClientMsgJoinRoom    = "JOIN_ROOM"
	ClientMsgLeaveRoom   = "LEAVE_ROOM"
	ClientMsgToggleReady = "TOGGLE_READY"
	ClientMsgFold        = "FOLD"
)

// Server-to-client message type constants.
const (
	ServerMsgPlayerJoined = "PLAYER_JOINED"
	ServerMsgPlayerLeft   = "PLAYER_LEFT"
	ServerMsgPlayerUpdate = "PLAYER_UPDATE"
	ServerMsgRoomState    = "ROOM_STATE"
	ServerMsgGameOver     = "GAME_OVER"
	ServerMsgError        = "ERROR"
)

// Envelope is the shared wire format for all WebSocket messages.
type Envelope[T any] struct {
	Type    string `json:"type"`
	Payload T      `json:"payload"`
}

// Player is the public player state transmitted over the wire.
type Player struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	IsHost   bool   `json:"isHost"`
	IsReady  bool   `json:"isReady"`
}

// JoinRoomPayload is sent by the client to join a room.
type JoinRoomPayload struct {
	RoomCode string `json:"roomCode"`
}

// PlayerLeftPayload is broadcast by the server when a player leaves or disconnects.
type PlayerLeftPayload struct {
	ID uint `json:"id"`
}

// RoomStatePayload is sent to a newly joined client with the current room roster.
type RoomStatePayload struct {
	RoomCode string   `json:"roomCode"`
	Players  []Player `json:"players"`
}

// GameOverPayload is broadcast when a hand is completed.
type GameOverPayload struct {
	WinnerID uint     `json:"winnerId"`
	Pot      int      `json:"pot"`
	Seats    []Player `json:"seats,omitempty"`
}

// ErrorPayload is sent by the server when a client message cannot be processed.
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
