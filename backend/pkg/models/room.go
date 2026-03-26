package models

import (
	"fmt"
	"regexp"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/utils"
	"gorm.io/gorm"
)

const (
	RoomStatusWaiting  = "WAITING"
	RoomStatusPlaying  = "PLAYING"
	RoomStatusFinished = "FINISHED"
)

var (
	roomCodePattern   = regexp.MustCompile(`^[A-Z0-9]{6}$`)
	allowedRoomStatus = map[string]struct{}{RoomStatusWaiting: {}, RoomStatusPlaying: {}, RoomStatusFinished: {}}
)

// Room represents a single poker game instance.
type Room struct {
	gorm.Model
	Code       string `gorm:"size:6;not null;uniqueIndex"`
	HostID     uint   `gorm:"not null;index"`
	Host       User   `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:HostID;references:ID"`
	Status     string `gorm:"type:text;not null;default:WAITING"`
	MaxPlayers int    `gorm:"not null"`
}

func (r *Room) BeforeSave(tx *gorm.DB) error {
	if r.Code == "" {
		code, err := utils.GenerateRoomCode()
		if err != nil {
			return err
		}
		r.Code = code
	}

	return r.validate()
}

func (r *Room) validate() error {
	if !roomCodePattern.MatchString(r.Code) {
		return fmt.Errorf("room code must be 6 uppercase alphanumeric characters")
	}

	if r.Status == "" {
		r.Status = RoomStatusWaiting
	}

	if _, ok := allowedRoomStatus[r.Status]; !ok {
		return fmt.Errorf("invalid room status %q", r.Status)
	}

	return nil
}
