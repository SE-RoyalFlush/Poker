package models

// AllModels returns a slice of all database models for auto-migration.
// Add new models to this slice to include them in the migration process.
func AllModels() []any {
	return []any{
		&User{},
		&Room{},
		&GameResult{},
	}
}
