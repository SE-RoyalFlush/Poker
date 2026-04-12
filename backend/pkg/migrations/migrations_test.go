package migrations_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/SE-RoyalFlush/Poker/backend/pkg/db"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/migrations"
	"github.com/SE-RoyalFlush/Poker/backend/pkg/models"
	"gorm.io/gorm/logger"
)

func TestRunMigrationsCreatesRoomsTable(t *testing.T) {
	db.ResetForTesting()
	t.Cleanup(func() {
		_ = db.Close()
	})

	database, err := db.Connect(&db.Config{
		DatabasePath:    filepath.Join(t.TempDir(), "migrations.db"),
		MaxOpenConns:    10,
		MaxIdleConns:    2,
		ConnMaxLifetime: time.Minute,
		LogLevel:        logger.Silent,
	})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	if err := migrations.RunMigrations(database); err != nil {
		t.Fatalf("RunMigrations returned error: %v", err)
	}

	if !database.Migrator().HasTable(&models.Room{}) {
		t.Fatal("expected rooms table to exist after migrations")
	}
}
