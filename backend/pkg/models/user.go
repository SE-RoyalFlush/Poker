package models

import (
	"gorm.io/gorm"
)

// User represents the data model for a system user.
type User struct {
	// gorm.Model includes ID, CreatedAt, UpdatedAt, and DeletedAt fields.
	gorm.Model

	// Username must be unique and is indexed for fast lookups.
	Username string `gorm:"uniqueIndex;not null" json:"username"`

	// PasswordHash stores the bcrypt hashed password.
	// The `json:"-"` tag ensures it is never serialized into JSON responses.
	PasswordHash string `gorm:"not null" json:"-"`
}
