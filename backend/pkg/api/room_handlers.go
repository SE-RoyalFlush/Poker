package api

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/auth"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	roomsvc "github.com/SE-RoyalFlush/Poker/backend/pkg/room"
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
	ID             uint   `json:"id"`
	Code           string `json:"code"`
	HostUserID     uint   `json:"hostUserId"`
	HostUsername   string `json:"hostUsername,omitempty"`
	Status         string `json:"status"`
	MaxPlayers     int    `json:"maxPlayers"`
	CurrentPlayers int    `json:"currentPlayers"`
	IsPrivate      bool   `json:"isPrivate"`
	IsFull         bool   `json:"isFull"`
	Seats          int    `json:"seats"`
}

func newRoomResponse(roomModel *models.Room) roomResponse {
	currentPlayers := globalWSHub.occupancy(roomModel.Code)
	seats := roomModel.MaxPlayers - currentPlayers
	if seats < 0 {
		seats = 0
	}

	return roomResponse{
		ID:             roomModel.ID,
		Code:           roomModel.Code,
		HostUserID:     roomModel.HostUserID,
		HostUsername:   roomModel.HostUser.Username,
		Status:         string(roomModel.Status),
		MaxPlayers:     roomModel.MaxPlayers,
		CurrentPlayers: currentPlayers,
		IsPrivate:      roomModel.IsPrivate,
		IsFull:         currentPlayers >= roomModel.MaxPlayers,
		Seats:          seats,
	}
}

// RoomsHandler handles GET /api/rooms requests.
func RoomsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	database, err := db.GetDB()
	if err != nil {
		sendError(w, "Internal Server Error", "Database not initialized", http.StatusInternalServerError)
		return
	}

	query := database.Preload("HostUser").Order("created_at DESC")
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	if status != "" {
		if !isValidRoomStatus(status) {
			sendError(w, "Bad Request", "Invalid room status", http.StatusBadRequest)
			return
		}
		query = query.Where("status = ?", status)
	}

	var rooms []models.Room
	if err := query.Find(&rooms).Error; err != nil {
		log.Printf("Failed to list rooms: %v", err)
		sendError(w, "Internal Server Error", "Failed to list rooms", http.StatusInternalServerError)
		return
	}

	response := make([]roomResponse, 0, len(rooms))
	for i := range rooms {
		response = append(response, newRoomResponse(&rooms[i]))
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode room list response: %v", err)
	}
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

	roomModel, err := roomsvc.Create(database, roomsvc.CreateParams{
		HostUserID: user.ID,
		MaxPlayers: req.MaxPlayers,
		IsPrivate:  req.IsPrivate,
	})
	if err != nil {
		if errors.Is(err, roomsvc.ErrInvalidMaxPlayers) {
			sendError(w, "Bad Request", "maxPlayers must be between 1 and 10", http.StatusBadRequest)
			return
		}
		if errors.Is(err, roomsvc.ErrFailedToGenerateCode) {
			sendError(w, "Internal Server Error", "Failed to generate room code", http.StatusInternalServerError)
			return
		}

		log.Printf("Failed to create room: %v", err)
		sendError(w, "Internal Server Error", "Failed to create room", http.StatusInternalServerError)
		return
	}
	roomModel.HostUser = *user

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(newRoomResponse(roomModel)); err != nil {
		log.Printf("Failed to encode room response: %v", err)
	}
}

// JoinRoomHandler handles POST /api/rooms/join requests.
func JoinRoomHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req joinRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Bad Request", "Invalid JSON input", http.StatusBadRequest)
		return
	}

	roomModel, ok := findOpenRoomForAPI(w, req.Code)
	if !ok {
		return
	}

	response := newRoomResponse(roomModel)
	if response.IsFull {
		sendError(w, "Forbidden", "Room is full", http.StatusForbidden)
		return
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode room response: %v", err)
	}
}

// RoomHandler handles GET /api/rooms/{code} requests.
func RoomHandler(w http.ResponseWriter, r *http.Request) {
	GetRoomHandler(w, r)
}

// GetRoomHandler handles GET /api/rooms/{code} requests.
func GetRoomHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	roomModel, ok := findOpenRoomForAPI(w, mux.Vars(r)["code"])
	if !ok {
		return
	}

	if err := json.NewEncoder(w).Encode(newRoomResponse(roomModel)); err != nil {
		log.Printf("Failed to encode room response: %v", err)
	}
}

func findOpenRoomForAPI(w http.ResponseWriter, code string) (*models.Room, bool) {
	database, err := db.GetDB()
	if err != nil {
		sendError(w, "Internal Server Error", "Database not initialized", http.StatusInternalServerError)
		return nil, false
	}

	roomModel, err := roomsvc.FindByCode(database, code)
	if err != nil {
		if errors.Is(err, roomsvc.ErrInvalidRoomCode) || errors.Is(err, gorm.ErrRecordNotFound) {
			sendError(w, "Not Found", "Room not found", http.StatusNotFound)
			return nil, false
		}

		log.Printf("Failed to fetch room by code: %v", err)
		sendError(w, "Internal Server Error", "Failed to fetch room", http.StatusInternalServerError)
		return nil, false
	}

	if roomModel.Status != models.RoomStatusOpen {
		sendError(w, "Not Found", "Room not found", http.StatusNotFound)
		return nil, false
	}

	return roomModel, true
}

func handleRoomAuthError(w http.ResponseWriter, err error) {
	if errors.Is(err, auth.ErrMissingSession) || errors.Is(err, auth.ErrInvalidSession) || errors.Is(err, auth.ErrUnauthenticated) {
		sendError(w, "Unauthorized", "Authentication required", http.StatusUnauthorized)
		return
	}

	sendError(w, "Internal Server Error", "Authentication failed", http.StatusInternalServerError)
}

func isValidRoomStatus(status string) bool {
	switch models.RoomStatus(status) {
	case models.RoomStatusOpen, models.RoomStatusClosed, models.RoomStatusInGame:
		return true
	default:
		return false
	}
}
