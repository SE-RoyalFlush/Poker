package utils

import (
	"crypto/rand"
	"fmt"
)

const roomCodeAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateRoomCode returns a random 6-character room invitation code.
func GenerateRoomCode() (string, error) {
	code := make([]byte, 6)
	randomBytes := make([]byte, 6)

	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("generate room code: %w", err)
	}

	for i, b := range randomBytes {
		code[i] = roomCodeAlphabet[int(b)%len(roomCodeAlphabet)]
	}

	return string(code), nil
}
