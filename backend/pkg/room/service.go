package room

import (
	"errors"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"gorm.io/gorm"
)

var (
	ErrInvalidMaxPlayers = errors.New("invalid max players")
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
	maxPlayers, err := validateMaxPlayers(params.MaxPlayers)
	if err != nil {
		return nil, err
	}

	room := &models.Room{
		HostUserID: params.HostUserID,
		Status:     params.Status,
		MaxPlayers: maxPlayers,
		IsPrivate:  params.IsPrivate,
	}

	if room.Status == "" {
		room.Status = models.RoomStatusOpen
	}

	if params.Code != "" {
		code, err := validateAndNormalizeCode(params.Code)
		if err != nil {
			return nil, err
		}
		room.Code = code

		if err := database.Create(room).Error; err != nil {
			return nil, err
		}

		return room, nil
	}

	return createWithGeneratedCode(database, room, GenerateRoomCode)
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

func createWithGeneratedCode(database *gorm.DB, room *models.Room, generator codeGenerator) (*models.Room, error) {
	for range maxRoomCodeCreateRetries {
		code, err := generateUniqueRoomCode(database, generator)
		if err != nil {
			if errors.Is(err, ErrFailedToGenerateCode) {
				continue
			}
			return nil, err
		}

		room.Code = code
		if err := database.Create(room).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) || isUniqueConstraintError(err) {
				continue
			}
			return nil, err
		}

		return room, nil
	}

	return nil, ErrFailedToGenerateCode
}

func validateMaxPlayers(maxPlayers int) (int, error) {
	if maxPlayers == 0 {
		return models.DefaultRoomMaxPlayers, nil
	}
	if maxPlayers < 0 || maxPlayers > models.MaxRoomPlayers {
		return 0, ErrInvalidMaxPlayers
	}

	return maxPlayers, nil
}
