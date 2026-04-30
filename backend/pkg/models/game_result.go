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
	DefaultGameType     GameType = GameTypeTexasHoldEm
)

// GameResult records the outcome of a completed poker hand.
type GameResult struct {
	gorm.Model

	WinnerID uint      `gorm:"not null;index" json:"winnerId"`
	Winner   User      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:WinnerID" json:"-"`
	PotSize  int       `gorm:"not null" json:"potSize"`
	Date     time.Time `gorm:"not null;index" json:"date"`
	GameType GameType  `gorm:"type:text;not null;default:texas_holdem;index" json:"gameType"`
}
