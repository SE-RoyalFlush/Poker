package api

import "net/http"

// WebSocketHandler is a protected placeholder for the authenticated websocket handshake flow.
func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	sendError(w, "Not Implemented", "WebSocket endpoint is not implemented yet", http.StatusNotImplemented)
}
