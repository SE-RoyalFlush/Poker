package models

import "gorm.io/gorm"

// RoomStatus represents the current lifecycle state of a poker room.
type RoomStatus string

const (
	RoomStatusOpen   RoomStatus = "open"
	RoomStatusClosed RoomStatus = "closed"
	RoomStatusInGame RoomStatus = "in_game"

	DefaultRoomMaxPlayers = 6
	MaxRoomPlayers        = 10
)

// Room represents a persisted poker room that can be created and looked up by a 6-character uppercase alphanumeric room code.
type Room struct {
	gorm.Model

	Code       string     `gorm:"size:6;uniqueIndex;not null"`
	HostUserID uint       `gorm:"not null;index"`
	HostUser   User       `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;foreignKey:HostUserID"`
	Status     RoomStatus `gorm:"type:text;not null;default:open;index"`
	MaxPlayers int        `gorm:"not null;default:6;check:max_players_range,max_players > 0 AND max_players <= 10"`
	IsPrivate  bool       `gorm:"not null;default:false"`
	IsActive   bool       `gorm:"not null;default:true;index"`
}
