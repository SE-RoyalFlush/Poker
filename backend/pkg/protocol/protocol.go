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
//	CHECK         {}
//	CALL          {}
//	RAISE         { amount: number }
//	CHAT_MESSAGE  { text: string }
//
// # Server → Client events
//
//	ROOM_STATE     { roomCode: string, players: Player[] }
//	PLAYER_JOINED  { id, username, isHost, isReady }
//	PLAYER_LEFT    { id }
//	PLAYER_UPDATE  { id, username, isHost, isReady }
//	GAME_STARTED   { tableId, seats, phase, pot, currentBet, activePlayerId, currentUserId }
//	CARDS_DEALT    { seats, holeCards }
//	PLAYER_ACTION  { playerId, action, amount, pot, currentBet, seats, activePlayerId }
//	PHASE_CHANGE   { phase, communityCards, pot, currentBet, activePlayerId }
//	GAME_OVER      { winnerId, pot, seats }
//	CHAT_MESSAGE   { senderId, username, text }
//	ERROR          { code: string, message: string }
package protocol

// Client-to-server message type constants.
const (
	ClientMsgJoinRoom    = "JOIN_ROOM"
	ClientMsgLeaveRoom   = "LEAVE_ROOM"
	ClientMsgToggleReady = "TOGGLE_READY"
	ClientMsgFold        = "FOLD"
	ClientMsgCheck       = "CHECK"
	ClientMsgCall        = "CALL"
	ClientMsgRaise       = "RAISE"
	ClientMsgChat        = "CHAT_MESSAGE"
)

// Server-to-client message type constants.
const (
	ServerMsgPlayerJoined = "PLAYER_JOINED"
	ServerMsgPlayerLeft   = "PLAYER_LEFT"
	ServerMsgPlayerUpdate = "PLAYER_UPDATE"
	ServerMsgRoomState    = "ROOM_STATE"
	ServerMsgGameStarted  = "GAME_STARTED"
	ServerMsgCardsDealt   = "CARDS_DEALT"
	ServerMsgPlayerAction = "PLAYER_ACTION"
	ServerMsgPhaseChange  = "PHASE_CHANGE"
	ServerMsgGameOver     = "GAME_OVER"
	ServerMsgChat         = "CHAT_MESSAGE"
	ServerMsgError        = "ERROR"
)

// Envelope is the shared wire format for all WebSocket messages.
type Envelope[T any] struct {
	Type    string `json:"type"`
	Payload T      `json:"payload"`
}

// Player is the public lobby player state transmitted over the wire.
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

// RaisePayload is sent by the client to raise.
type RaisePayload struct {
	Amount int `json:"amount"`
}

// ChatInPayload is sent by the client to send a chat message.
type ChatInPayload struct {
	Text string `json:"text"`
}

// ChatPayload is broadcast by the server with a chat message.
type ChatPayload struct {
	SenderID uint   `json:"senderId"`
	Username string `json:"username"`
	Text     string `json:"text"`
}

// ErrorPayload is sent by the server when a client message cannot be processed.
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
