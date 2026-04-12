package middleware

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/auth"
)

type errorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

// AuthMiddleware validates the signed session cookie and resolves the authenticated user.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := auth.AuthenticatedUserFromRequest(r)
		if err != nil {
			if errors.Is(err, auth.ErrMissingSession) || errors.Is(err, auth.ErrInvalidSession) || errors.Is(err, auth.ErrUnauthenticated) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_ = json.NewEncoder(w).Encode(errorResponse{
					Error:   "Unauthorized",
					Message: "Authentication required",
					Status:  http.StatusUnauthorized,
				})
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(errorResponse{
				Error:   "Internal Server Error",
				Message: "Authentication failed",
				Status:  http.StatusInternalServerError,
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}
