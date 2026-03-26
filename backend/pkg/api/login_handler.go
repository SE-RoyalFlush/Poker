package api

import (
	"encoding/json"
	"net/http"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/auth"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/middleware"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
)

// LoginRequest represents the login request payload
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginHandler handles user login and sets JWT in HttpOnly cookie
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		sendError(w, "Method Not Allowed", "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Bad Request", "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		sendError(w, "Bad Request", "Username and password required", http.StatusBadRequest)
		return
	}

	// Fetch user from database
	database, err := db.GetDB()
	if err != nil {
		sendError(w, "Internal Server Error", "Database error", http.StatusInternalServerError)
		return
	}

	var user models.User
	if err := database.Where("username = ?", req.Username).First(&user).Error; err != nil {
		// Invalid credentials - don't reveal if user exists
		sendError(w, "Unauthorized", "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Verify password
	if err := auth.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		// Invalid credentials
		sendError(w, "Unauthorized", "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(&user, 24)
	if err != nil {
		sendError(w, "Internal Server Error", "Token generation failed", http.StatusInternalServerError)
		return
	}

	// Determine whether to set the Secure flag based on TLS.
	secureCookie := r.TLS != nil

	// Set HttpOnly, SameSite cookie with conditional Secure flag
	cookie := &http.Cookie{
		Name:     middleware.CookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   24 * 60 * 60, // 24 hours
		HttpOnly: true,
		Secure:   secureCookie, // Only sent over HTTPS when TLS is enabled
		SameSite: http.SameSiteStrictMode,
	}
	http.SetCookie(w, cookie)

	// Return user info (NOT the token)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "login successful",
		"user":    user.ToResponse(),
	})
}
