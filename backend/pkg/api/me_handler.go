package api

import (
	"encoding/json"
	"net/http"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/auth"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/middleware"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
)

// MeHandler authenticates the user via the JWT cookie and returns the current user's info.
// Authentication is performed within this handler using the token from the HttpOnly cookie.
func MeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, `{"error": "method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	// Get the JWT token from HttpOnly cookie
	cookie, err := r.Cookie(middleware.CookieName)
	if err != nil {
		if err == http.ErrNoCookie {
			http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
			return
		}
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Verify the token
	claims, err := auth.VerifyToken(cookie.Value)
	if err != nil {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	response := models.UserResponse{
		ID:        claims.UserID,
		Username:  claims.Username,
		CreatedAt: claims.CreatedAt,
		UpdatedAt: claims.UpdatedAt,
	}

	// Support older tokens that do not yet include the full profile payload.
	// New tokens avoid this database round-trip entirely.
	if response.Username == "" || response.CreatedAt.IsZero() || response.UpdatedAt.IsZero() {
		database, err := db.GetDB()
		if err != nil {
			http.Error(w, `{"error": "database error"}`, http.StatusInternalServerError)
			return
		}

		var user models.User
		if err := database.First(&user, claims.UserID).Error; err != nil {
			http.Error(w, `{"error": "user not found"}`, http.StatusNotFound)
			return
		}

		response = user.ToResponse()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
