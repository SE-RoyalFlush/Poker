package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// RegisterRequest defines the input for the registration endpoint.
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterHandler handles POST /api/register requests.
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Bad Request", "Invalid JSON input", http.StatusBadRequest)
		return
	}

	// Validation: Username non-empty and min length 3
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || len(req.Username) < 3 {
		sendError(w, "Bad Request", "Username must be at least 3 characters", http.StatusBadRequest)
		return
	}
	// Validation: Password min length 6
	if len(req.Password) < 6 {
		sendError(w, "Bad Request", "Password must be at least 6 characters", http.StatusBadRequest)
		return
	}

	database, err := db.GetDB()
	if err != nil {
		sendError(w, "Internal Server Error", "Database not initialized", http.StatusInternalServerError)
		return
	}

	// Duplicate Check: Explicitly check for an existing username to return 409 Conflict
	var existingUser models.User
	err = database.Where("username = ?", req.Username).First(&existingUser).Error
	if err == nil {
		sendError(w, "Conflict", "Username already exists", http.StatusConflict)
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Printf("Database error during duplicate check: %v", err)
		sendError(w, "Internal Server Error", "Database error", http.StatusInternalServerError)
		return
	}

	// Hashing: Use bcrypt to securely hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password: %v", err)
		sendError(w, "Internal Server Error", "Failed to process password", http.StatusInternalServerError)
		return
	}

	// Persistence: Create the new user in the database
	newUser := models.User{
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
	}

	if err := database.Create(&newUser).Error; err != nil {
		log.Printf("Failed to create user: %v", err)
		sendError(w, "Internal Server Error", "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Success Response: Return 201 Created and the user object (hash excluded by JSON tag)
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(newUser); err != nil {
		log.Printf("Failed to encode user response: %v", err)
	}
}

// sendError is a helper to return JSON-formatted error responses.
func sendError(w http.ResponseWriter, errType, message string, status int) {
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(ErrorResponse{
		Error:   errType,
		Message: message,
		Status:  status,
	}); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}
