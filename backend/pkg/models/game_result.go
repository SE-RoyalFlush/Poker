package models

import (
	"time"

	"gorm.io/gorm"
)

// GameType represents the variant of poker played in a game.
type GameType string

const (
	GameTypeTexasHoldEm GameType = "texas_holdem"
	GameTypeOmaha       GameType = "omaha"
)

// GameResult records the outcome of a completed poker game.
type GameResult struct {
	gorm.Model

	WinnerID uint     `gorm:"not null;index"`
	Winner   User     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:WinnerID"`
	PotSize  int      `gorm:"not null"`
	Date     time.Time `gorm:"not null;index"`
	GameType GameType  `gorm:"type:text;not null;default:texas_holdem;index"`
}
