package models

import "gorm.io/gorm"

// User is the minimal account model required for relational ownership.
type User struct {
	gorm.Model
}
