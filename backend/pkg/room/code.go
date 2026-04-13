package room

import (
	"crypto/rand"
	"errors"
	"regexp"
	"strings"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"gorm.io/gorm"
)

const (
	RoomCodeLength           = 6
	roomCodeAlphabet         = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	maxRoomCodeCreateRetries = 5
)

var (
	ErrInvalidRoomCode      = errors.New("invalid room code")
	ErrFailedToGenerateCode = errors.New("failed to generate unique room code")
	roomCodePattern         = regexp.MustCompile(`^[A-Z0-9]{6}$`)
)

type codeGenerator func() (string, error)

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

func GenerateRoomCode() (string, error) {
	buffer := make([]byte, RoomCodeLength)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	for i := range buffer {
		buffer[i] = roomCodeAlphabet[int(buffer[i])%len(roomCodeAlphabet)]
	}

	return string(buffer), nil
}

func generateUniqueRoomCode(database *gorm.DB, generator codeGenerator) (string, error) {
	for range maxRoomCodeCreateRetries {
		code, err := generator()
		if err != nil {
			return "", err
		}

		normalized, err := validateAndNormalizeCode(code)
		if err != nil {
			return "", err
		}

		var count int64
		if err := database.Model(&models.Room{}).Where("code = ?", normalized).Count(&count).Error; err != nil {
			return "", err
		}
		if count == 0 {
			return normalized, nil
		}
	}

	return "", ErrFailedToGenerateCode
}

func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(strings.ToLower(err.Error()), "unique constraint failed")
}
