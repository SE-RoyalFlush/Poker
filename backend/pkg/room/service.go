package room

import (
	"strings"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"gorm.io/gorm"
)

const defaultMaxPlayers = 6

// CreateParams defines the persisted room attributes required for creation.
type CreateParams struct {
	Code       string
	HostUserID uint
	Status     models.RoomStatus
	MaxPlayers int
	IsPrivate  bool
}

// Create inserts a room row and returns the persisted record.
func Create(database *gorm.DB, params CreateParams) (*models.Room, error) {
	room := &models.Room{
		Code:       normalizeCode(params.Code),
		HostUserID: params.HostUserID,
		Status:     params.Status,
		MaxPlayers: params.MaxPlayers,
		IsPrivate:  params.IsPrivate,
	}

	if room.Status == "" {
		room.Status = models.RoomStatusOpen
	}
	if room.MaxPlayers == 0 {
		room.MaxPlayers = defaultMaxPlayers
	}

	if err := database.Create(room).Error; err != nil {
		return nil, err
	}

	return room, nil
}

// FindByCode retrieves a room by its invite code.
func FindByCode(database *gorm.DB, code string) (*models.Room, error) {
	var room models.Room
	if err := database.Preload("HostUser").Where("code = ?", normalizeCode(code)).First(&room).Error; err != nil {
		return nil, err
	}

	return &room, nil
}

func normalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}
