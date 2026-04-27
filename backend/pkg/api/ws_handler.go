package api

import (
	"net/http"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/auth"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/socket"
)

// WebSocketHandler upgrades the authenticated request and manages lobby events.
func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	user, err := auth.AuthenticatedUserFromRequest(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		sendError(w, "Unauthorized", "Authentication required", http.StatusUnauthorized)
		return
	}

	socket.ServeWs(w, r, user)
}
