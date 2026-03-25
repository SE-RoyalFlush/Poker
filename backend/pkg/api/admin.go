package api

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"github.com/gorilla/mux"
)

type adminUserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

func adminCredentials() (string, string) {
	username := os.Getenv("ADMIN_USERNAME")
	password := os.Getenv("ADMIN_PASSWORD")

	if username == "" {
		username = "admin"
	}
	if password == "" {
		password = "admin"
	}

	return username, password
}

func requireAdminAuth(w http.ResponseWriter, r *http.Request) bool {
	expectedUser, expectedPassword := adminCredentials()
	username, password, ok := r.BasicAuth()
	if !ok || username != expectedUser || password != expectedPassword {
		w.Header().Set("WWW-Authenticate", `Basic realm="admin"`)
		sendError(w, "Unauthorized", "Admin credentials required", http.StatusUnauthorized)
		return false
	}

	return true
}

// AdminListUsersHandler handles GET /api/admin/users requests.
func AdminListUsersHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if !requireAdminAuth(w, r) {
		return
	}

	database, err := db.GetDB()
	if err != nil {
		sendError(w, "Internal Server Error", "Database not initialized", http.StatusInternalServerError)
		return
	}

	var users []models.User
	if err := database.Order("id ASC").Find(&users).Error; err != nil {
		sendError(w, "Internal Server Error", "Failed to fetch users", http.StatusInternalServerError)
		return
	}

	response := make([]adminUserResponse, 0, len(users))
	for _, user := range users {
		response = append(response, adminUserResponse{
			ID:       user.ID,
			Username: user.Username,
		})
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		sendError(w, "Internal Server Error", "Failed to encode user list", http.StatusInternalServerError)
		return
	}
}

// AdminDeleteUserHandler handles DELETE /api/admin/users/{id} requests.
func AdminDeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if !requireAdminAuth(w, r) {
		return
	}

	database, err := db.GetDB()
	if err != nil {
		sendError(w, "Internal Server Error", "Database not initialized", http.StatusInternalServerError)
		return
	}

	idStr := mux.Vars(r)["id"]
	userID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		sendError(w, "Bad Request", "Invalid user id", http.StatusBadRequest)
		return
	}

	result := database.Delete(&models.User{}, userID)
	if result.Error != nil {
		sendError(w, "Internal Server Error", "Failed to delete user", http.StatusInternalServerError)
		return
	}
	if result.RowsAffected == 0 {
		sendError(w, "Not Found", "User not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
