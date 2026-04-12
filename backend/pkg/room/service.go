package room

import (
	"errors"
	"regexp"
	"strings"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"gorm.io/gorm"
)

const defaultMaxPlayers = 6
const maxAllowedPlayers = 10

var (
	ErrInvalidRoomCode   = errors.New("invalid room code")
	ErrInvalidMaxPlayers = errors.New("invalid max players")
	roomCodePattern      = regexp.MustCompile(`^[A-Z0-9]{6}$`)
)

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
	code, err := validateAndNormalizeCode(params.Code)
	if err != nil {
		return nil, err
	}
	maxPlayers, err := validateMaxPlayers(params.MaxPlayers)
	if err != nil {
		return nil, err
	}

	room := &models.Room{
		Code:       code,
		HostUserID: params.HostUserID,
		Status:     params.Status,
		MaxPlayers: maxPlayers,
		IsPrivate:  params.IsPrivate,
	}

	if room.Status == "" {
		room.Status = models.RoomStatusOpen
	}

	if err := database.Create(room).Error; err != nil {
		return nil, err
	}

	return room, nil
}

// FindByCode retrieves a room by its invite code.
func FindByCode(database *gorm.DB, code string) (*models.Room, error) {
	normalizedCode, err := validateAndNormalizeCode(code)
	if err != nil {
		return nil, err
	}

	var room models.Room
	if err := database.Preload("HostUser").Where("code = ?", normalizedCode).First(&room).Error; err != nil {
		return nil, err
	}

	return &room, nil
}

func normalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}

func validateAndNormalizeCode(code string) (string, error) {
	normalized := normalizeCode(code)
	if !roomCodePattern.MatchString(normalized) {
		return "", ErrInvalidRoomCode
	}

	return normalized, nil
}

func validateMaxPlayers(maxPlayers int) (int, error) {
	if maxPlayers == 0 {
		return defaultMaxPlayers, nil
	}
	if maxPlayers < 0 || maxPlayers > maxAllowedPlayers {
		return 0, ErrInvalidMaxPlayers
	}

	return maxPlayers, nil
}
