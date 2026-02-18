package api

import (
	"encoding/json"
	"log"
	"net/http"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func NotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)

	if err := json.NewEncoder(w).Encode(ErrorResponse{
		Error:   "Not Found",
		Message: "The requested resource was not found",
		Status:  http.StatusNotFound,
	}); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}
