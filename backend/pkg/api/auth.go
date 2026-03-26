package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"github.com/gorilla/securecookie"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	// Session cookie name
	sessionCookieName = "session-id"

	// sessionKey should be 32 or 64 bytes. In production, load this from an environment variable.
	sessionKey = []byte(os.Getenv("SESSION_KEY"))

	// secureCookie is used to sign and encrypt cookies.
	secureCookie *securecookie.SecureCookie
)

func init() {
	if len(sessionKey) == 0 {
		// Fallback for development only
		sessionKey = []byte("dev-only-32-byte-session-secret-")
	}
	secureCookie = securecookie.New(sessionKey, nil)
}

// MeHandler handles GET /api/me requests to return the current authenticated user.
// Returns 204 No Content if not authenticated (instead of 401) to allow frontend silent session check.
func MeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var username string
	if err := secureCookie.Decode(sessionCookieName, cookie.Value, &username); err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if username == "" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	database, err := db.GetDB()
	if err != nil {
		sendError(w, "Internal Server Error", "Database not initialized", http.StatusInternalServerError)
		return
	}

	var user models.User
	if err := database.Where("username = ?", username).First(&user).Error; err != nil {
		// If credentials provided but user not found, return 401
		sendError(w, "Unauthorized", "User not found", http.StatusUnauthorized)
		return
	}

	// Return user info (password hash is omitted by JSON tag)
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		log.Printf("Failed to encode user response: %v", err)
	}
}

// LoginRequest defines the input for the login endpoint.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginHandler handles POST /api/login requests.
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Bad Request", "Invalid JSON input", http.StatusBadRequest)
		return
	}

	database, err := db.GetDB()
	if err != nil {
		sendError(w, "Internal Server Error", "Database not initialized", http.StatusInternalServerError)
		return
	}

	var user models.User
	if err := database.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			sendError(w, "Unauthorized", "Invalid username or password", http.StatusUnauthorized)
		} else {
			sendError(w, "Internal Server Error", "Database error", http.StatusInternalServerError)
		}
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		sendError(w, "Unauthorized", "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Create session cookie
	encoded, err := secureCookie.Encode(sessionCookieName, user.Username)
	if err != nil {
		log.Printf("Failed to encode session cookie: %v", err)
		sendError(w, "Internal Server Error", "Failed to create session", http.StatusInternalServerError)
		return
	}

	// Set cookie with appropriate security settings
	appEnv := os.Getenv("APP_ENV")
	goEnv := os.Getenv("GO_ENV")
	secure := true
	if appEnv == "development" || goEnv == "development" {
		secure = false
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    encoded,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusNoContent)
}

// LogoutHandler handles POST /api/logout requests.
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	w.WriteHeader(http.StatusNoContent)
}

func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}

	// SQLite duplicate key errors are usually surfaced as raw messages.
	return strings.Contains(strings.ToLower(err.Error()), "unique constraint failed")
}

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
		// Handle unique constraint violations (e.g., concurrent registration with same username)
		if errors.Is(err, gorm.ErrDuplicatedKey) || isUniqueConstraintError(err) {
			sendError(w, "Conflict", "Username already exists", http.StatusConflict)
			return
		}
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
