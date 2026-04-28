package api

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/auth"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/room"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

type createRoomRequest struct {
	MaxPlayers int  `json:"maxPlayers"`
	IsPrivate  bool `json:"isPrivate"`
}

type joinRoomRequest struct {
	Code string `json:"code"`
}

type roomResponse struct {
	ID             string            `json:"id"`
	Code           string            `json:"code"`
	Name           string            `json:"name"`
	GameType       string            `json:"gameType"`
	SmallBlind     int               `json:"smallBlind"`
	BigBlind       int               `json:"bigBlind"`
	MaxPlayers     int               `json:"maxPlayers"`
	CurrentPlayers int               `json:"currentPlayers"`
	IsPrivate      bool              `json:"isPrivate"`
	IsFull         bool              `json:"isFull"`
	Seats          int               `json:"seats"`
	Status         models.RoomStatus `json:"status"`
	HostUserID     uint              `json:"hostUserId"`
	HostUsername   string            `json:"hostUsername,omitempty"`
}

// RoomsHandler handles GET /api/rooms requests.
func RoomsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	listRooms(w, r)
}

// CreateRoomHandler handles POST /api/rooms requests.
func CreateRoomHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	user, err := auth.AuthenticatedUserFromRequest(r)
	if err != nil {
		handleRoomAuthError(w, err)
		return
	}

	var req createRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		sendError(w, "Bad Request", "Invalid JSON input", http.StatusBadRequest)
		return
	}

	database, err := db.GetDB()
	if err != nil {
		sendError(w, "Internal Server Error", "Database not initialized", http.StatusInternalServerError)
		return
	}

	roomModel, err := room.Create(database, room.CreateParams{
		HostUserID: user.ID,
		MaxPlayers: req.MaxPlayers,
		IsPrivate:  req.IsPrivate,
	})
	if err != nil {
		if errors.Is(err, room.ErrInvalidMaxPlayers) {
			sendError(w, "Bad Request", "maxPlayers must be between 1 and 10", http.StatusBadRequest)
			return
		}
		if errors.Is(err, room.ErrFailedToGenerateCode) {
			sendError(w, "Internal Server Error", "Failed to generate room code", http.StatusInternalServerError)
			return
		}

		log.Printf("Failed to create room: %v", err)
		sendError(w, "Internal Server Error", "Failed to create room", http.StatusInternalServerError)
		return
	}
	roomModel.HostUser = *user

	writeRoomResponse(w, http.StatusCreated, roomModel)
}

// JoinRoomHandler verifies a room exists and is open. Active socket membership
// is established later by JOIN_ROOM over the WebSocket connection.
func JoinRoomHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req joinRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Bad Request", "Invalid JSON input", http.StatusBadRequest)
		return
	}

	roomModel, err := findPersistedRoom(req.Code)
	if err != nil {
		handleRoomLookupError(w, err)
		return
	}
	if roomModel.Status != models.RoomStatusOpen {
		sendError(w, "Not Found", "Room not found", http.StatusNotFound)
		return
	}

	response := buildRoomResponse(roomModel)
	if response.IsFull {
		sendError(w, "Forbidden", "Room is full", http.StatusForbidden)
		return
	}

	writeRoomResponse(w, http.StatusOK, roomModel)
}

// RoomHandler handles GET /api/rooms/{code} requests.
func RoomHandler(w http.ResponseWriter, r *http.Request) {
	GetRoomHandler(w, r)
}

// GetRoomHandler looks up persisted room metadata by code without requiring any
// active WebSocket connection for that room.
func GetRoomHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	roomModel, err := findPersistedRoom(mux.Vars(r)["code"])
	if err != nil {
		handleRoomLookupError(w, err)
		return
	}
	if roomModel.Status != models.RoomStatusOpen {
		sendError(w, "Not Found", "Room not found", http.StatusNotFound)
		return
	}

	writeRoomResponse(w, http.StatusOK, roomModel)
}

func listRooms(w http.ResponseWriter, r *http.Request) {
	database, err := db.GetDB()
	if err != nil {
		sendError(w, "Internal Server Error", "Database not initialized", http.StatusInternalServerError)
		return
	}

	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status != "" && !room.IsValidStatus(models.RoomStatus(status)) {
		sendError(w, "Bad Request", "Invalid room status", http.StatusBadRequest)
		return
	}

	rooms, err := room.List(database, room.ListParams{
		Status: models.RoomStatus(status),
	})
	if err != nil {
		log.Printf("Failed to list rooms: %v", err)
		sendError(w, "Internal Server Error", "Failed to list rooms", http.StatusInternalServerError)
		return
	}

	response := make([]roomResponse, 0, len(rooms))
	for i := range rooms {
		response = append(response, buildRoomResponse(&rooms[i]))
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode room list response: %v", err)
	}
}

func findPersistedRoom(code string) (*models.Room, error) {
	database, err := db.GetDB()
	if err != nil {
		return nil, err
	}

	return room.FindByCode(database, code)
}

func handleRoomAuthError(w http.ResponseWriter, err error) {
	if errors.Is(err, auth.ErrMissingSession) || errors.Is(err, auth.ErrInvalidSession) || errors.Is(err, auth.ErrUnauthenticated) {
		sendError(w, "Unauthorized", "Authentication required", http.StatusUnauthorized)
		return
	}

	sendError(w, "Internal Server Error", "Authentication failed", http.StatusInternalServerError)
}

func handleRoomLookupError(w http.ResponseWriter, err error) {
	if errors.Is(err, room.ErrInvalidRoomCode) {
		sendError(w, "Bad Request", "Invalid room code", http.StatusBadRequest)
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		sendError(w, "Not Found", "Room not found", http.StatusNotFound)
		return
	}
	if errors.Is(err, db.ErrNotInitialized) {
		sendError(w, "Internal Server Error", "Database not initialized", http.StatusInternalServerError)
		return
	}

	log.Printf("Failed to look up room: %v", err)
	sendError(w, "Internal Server Error", "Database error", http.StatusInternalServerError)
}

func writeRoomResponse(w http.ResponseWriter, status int, roomModel *models.Room) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(buildRoomResponse(roomModel)); err != nil {
		log.Printf("Failed to encode room response: %v", err)
	}
}

func buildRoomResponse(roomModel *models.Room) roomResponse {
	currentPlayers := globalWSHub.activePlayerCount(roomModel.Code)
	if currentPlayers > roomModel.MaxPlayers {
		log.Printf("WARNING: room %s hub player count (%d) exceeds maxPlayers (%d); capping for response", roomModel.Code, currentPlayers, roomModel.MaxPlayers)
		currentPlayers = roomModel.MaxPlayers
	}
	seats := roomModel.MaxPlayers - currentPlayers

	return roomResponse{
		ID:             strconv.FormatUint(uint64(roomModel.ID), 10),
		Code:           roomModel.Code,
		Name:           "Room " + roomModel.Code,
		GameType:       "NLH",
		SmallBlind:     1,
		BigBlind:       2,
		MaxPlayers:     roomModel.MaxPlayers,
		CurrentPlayers: currentPlayers,
		IsPrivate:      roomModel.IsPrivate,
		IsFull:         currentPlayers >= roomModel.MaxPlayers,
		Seats:          seats,
		Status:         roomModel.Status,
		HostUserID:     roomModel.HostUserID,
		HostUsername:   roomModel.HostUser.Username,
	}
}

