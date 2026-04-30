package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
)

// LeaderboardEntry is one row in the leaderboard response.
type LeaderboardEntry struct {
	Rank          int     `json:"rank"`
	UserID        uint    `json:"userId"`
	Username      string  `json:"username"`
	HandsPlayed   int64   `json:"handsPlayed"`
	Wins          int64   `json:"wins"`
	WinRate       float64 `json:"winRate"`
	TotalEarnings int64   `json:"totalEarnings"`
}

type leaderboardRow struct {
	ID            uint
	Username      string
	Wins          int64
	TotalEarnings int64
}

// LeaderboardHandler handles GET /api/leaderboard requests.
func LeaderboardHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	database, err := db.GetDB()
	if err != nil {
		sendError(w, "Internal Server Error", "Database not initialized", http.StatusInternalServerError)
		return
	}

	var rows []leaderboardRow
	if err := database.Model(&models.User{}).
		Select("users.id, users.username, COUNT(game_results.id) as wins, COALESCE(SUM(game_results.pot_size), 0) as total_earnings").
		Joins("LEFT JOIN game_results ON game_results.winner_id = users.id AND game_results.deleted_at IS NULL").
		Group("users.id").
		Order("wins DESC").
		Limit(50).
		Scan(&rows).Error; err != nil {
		log.Printf("Failed to query leaderboard: %v", err)
		sendError(w, "Internal Server Error", "Database error", http.StatusInternalServerError)
		return
	}

	entries := make([]LeaderboardEntry, len(rows))
	for i, row := range rows {
		var winRate float64
		if row.Wins > 0 {
			winRate = 100.0
		}
		entries[i] = LeaderboardEntry{
			Rank:          i + 1,
			UserID:        row.ID,
			Username:      row.Username,
			HandsPlayed:   row.Wins,
			Wins:          row.Wins,
			WinRate:       winRate,
			TotalEarnings: row.TotalEarnings,
		}
	}

	if err := json.NewEncoder(w).Encode(entries); err != nil {
		log.Printf("Failed to encode leaderboard response: %v", err)
	}
}
