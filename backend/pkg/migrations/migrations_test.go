package migrations_test

import (
	"path/filepath"
	"testing"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/migrations"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRunMigrationsCreatesRoomsTable(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "migrations.db")
	database, err := gorm.Open(sqlite.Open(databasePath), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatalf("failed to unwrap sql database: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if database.Migrator().HasTable(&models.Room{}) {
		t.Fatal("expected rooms table to not exist before migrations")
	}

	if err := migrations.RunMigrations(database); err != nil {
		t.Fatalf("RunMigrations returned error: %v", err)
	}

	if !database.Migrator().HasTable(&models.Room{}) {
		t.Fatal("expected rooms table to exist after migrations")
	}
}
