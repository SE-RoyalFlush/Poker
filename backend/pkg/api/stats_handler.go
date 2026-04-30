package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

// UserStatsResponse is the aggregate statistics payload for a user profile.
type UserStatsResponse struct {
	HandsPlayed   int64   `json:"handsPlayed"`
	Wins          int64   `json:"wins"`
	Losses        int64   `json:"losses"`
	WinRate       float64 `json:"winRate"`
	TotalEarnings int64   `json:"totalEarnings"`
}

// UserStatsHandler handles GET /api/users/{id}/stats requests.
func UserStatsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userID, err := strconv.ParseUint(mux.Vars(r)["id"], 10, 64)
	if err != nil || userID == 0 {
		sendError(w, "Bad Request", "Invalid user ID", http.StatusBadRequest)
		return
	}

	database, err := db.GetDB()
	if err != nil {
		sendError(w, "Internal Server Error", "Database not initialized", http.StatusInternalServerError)
		return
	}

	var user models.User
	if err := database.First(&user, uint(userID)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			sendError(w, "Not Found", "User not found", http.StatusNotFound)
			return
		}

		log.Printf("Failed to look up user %d for stats: %v", userID, err)
		sendError(w, "Internal Server Error", "Database error", http.StatusInternalServerError)
		return
	}

	stats, err := buildUserStats(database, user.ID)
	if err != nil {
		log.Printf("Failed to aggregate stats for user %d: %v", user.ID, err)
		sendError(w, "Internal Server Error", "Database error", http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(stats); err != nil {
		log.Printf("Failed to encode stats response: %v", err)
	}
}

func buildUserStats(database *gorm.DB, userID uint) (UserStatsResponse, error) {
	var handsPlayed int64
	if err := database.Model(&models.GameResult{}).Count(&handsPlayed).Error; err != nil {
		return UserStatsResponse{}, err
	}

	var wins int64
	if err := database.Model(&models.GameResult{}).Where("winner_id = ?", userID).Count(&wins).Error; err != nil {
		return UserStatsResponse{}, err
	}

	winnings, err := sumPots(database.Where("winner_id = ?", userID))
	if err != nil {
		return UserStatsResponse{}, err
	}

	losses := handsPlayed - wins
	lossesPotTotal, err := sumPots(database.Where("winner_id <> ?", userID))
	if err != nil {
		return UserStatsResponse{}, err
	}

	var winRate float64
	if handsPlayed > 0 {
		winRate = (float64(wins) / float64(handsPlayed)) * 100
	}

	return UserStatsResponse{
		HandsPlayed:   handsPlayed,
		Wins:          wins,
		Losses:        losses,
		WinRate:       winRate,
		TotalEarnings: winnings - lossesPotTotal,
	}, nil
}

func sumPots(query *gorm.DB) (int64, error) {
	var total int64
	if err := query.Model(&models.GameResult{}).Select("COALESCE(SUM(pot_size), 0)").Scan(&total).Error; err != nil {
		return 0, err
	}
	return total, nil
}
