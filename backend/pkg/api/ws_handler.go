package api

import "net/http"

// WebSocketHandler is a protected placeholder for the authenticated websocket handshake flow.
func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	sendError(w, "Not Implemented", "WebSocket endpoint is not implemented yet", http.StatusNotImplemented)
}
