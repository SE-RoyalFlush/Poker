package api

import "net/http"

// RoomsHandler is a protected placeholder for room collection endpoints until room features are implemented.
func RoomsHandler(w http.ResponseWriter, r *http.Request) {
	sendError(w, "Not Implemented", "Room endpoints are not implemented yet", http.StatusNotImplemented)
}

// JoinRoomHandler is a protected placeholder for room join flows until room features are implemented.
func JoinRoomHandler(w http.ResponseWriter, r *http.Request) {
	sendError(w, "Not Implemented", "Room endpoints are not implemented yet", http.StatusNotImplemented)
}

// RoomHandler is a protected placeholder for room lookup flows until room features are implemented.
func RoomHandler(w http.ResponseWriter, r *http.Request) {
	sendError(w, "Not Implemented", "Room endpoints are not implemented yet", http.StatusNotImplemented)
}
