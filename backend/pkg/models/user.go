package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents the data model for a system user.
type User struct {
	gorm.Model

	Username     string `gorm:"uniqueIndex;not null" json:"username"`
	PasswordHash string `gorm:"not null" json:"-"`
}

// UserResponse returns the public user JSON shape expected by the frontend auth flow.
type UserResponse struct {
	ID        uint      `json:"ID"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"CreatedAt"`
	UpdatedAt time.Time `json:"UpdatedAt"`
}

// ToResponse returns a user payload without sensitive fields.
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
