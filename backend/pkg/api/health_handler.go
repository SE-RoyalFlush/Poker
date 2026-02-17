package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
)

type HealthResponse struct {
	Status   string `json:"status"`
	Database string `json:"database"`
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := HealthResponse{
		Status:   "alive",
		Database: "disconnected",
	}

	// Check database connection
	if err := db.Ping(); err == nil {
		response.Database = "connected"
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode health response: %v", err)
	}
}
