package db

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	instance *gorm.DB
	once     sync.Once
	dbErr    error

	ErrNotInitialized = errors.New("database not initialized: call Connect() first")
)

// Config holds database configuration
type Config struct {
	DatabasePath    string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	LogLevel        logger.LogLevel
}

// DefaultConfig returns sensible defaults for development
func DefaultConfig() *Config {
	return &Config{
		DatabasePath:    "./data/poker.db",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		LogLevel:        logger.Info,
	}
}

// Connect initializes and returns the singleton database instance
// Uses sync.Once to ensure thread-safe single initialization
func Connect(cfg *Config) (*gorm.DB, error) {
	once.Do(func() {
		if cfg == nil {
			cfg = DefaultConfig()
		}

		if os.Getenv("ENVIRONMENT") == "production" {
			cfg.LogLevel = logger.Silent
		}

		// Ensure the data directory exists
		dir := filepath.Dir(cfg.DatabasePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			dbErr = err
			log.Printf("Failed to create database directory: %v", err)
			return
		}

		gormLogger := logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  cfg.LogLevel,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
				ParameterizedQueries:      false,
			},
		)

		instance, dbErr = gorm.Open(sqlite.Open(cfg.DatabasePath), &gorm.Config{
			Logger: gormLogger,
		})

		if dbErr != nil {
			log.Printf("Failed to connect to database: %v", dbErr)
			return
		}

		sqlDB, err := instance.DB()
		if err != nil {
			dbErr = err
			log.Printf("Failed to get database instance: %v", err)
			return
		}

		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

		if err := instance.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
			dbErr = err
			log.Printf("Failed to enable foreign keys: %v", err)
			return
		}

		var fkEnabled int
		if err := instance.Raw("PRAGMA foreign_keys").Scan(&fkEnabled).Error; err != nil {
			log.Printf("Warning: Could not verify foreign keys status: %v", err)
		} else if fkEnabled == 1 {
			log.Println("✓ Foreign keys enabled")
		}

		log.Printf("✓ Database connection established: %s", cfg.DatabasePath)
	})

	return instance, dbErr
}

// GetDB returns the existing database instance
// Returns error if Connect() hasn't been called successfully
func GetDB() (*gorm.DB, error) {
	if instance == nil {
		if dbErr != nil {
			return nil, dbErr
		}
		return nil, ErrNotInitialized
	}
	return instance, nil
}

// Close closes the database connection
// Should be called during graceful shutdown
func Close() error {
	if instance == nil {
		return nil
	}

	sqlDB, err := instance.DB()
	if err != nil {
		return err
	}

	if err := sqlDB.Close(); err != nil {
		log.Printf("Error closing database: %v", err)
		return err
	}

	log.Println("✓ Database connection closed")
	return nil
}

// Ping verifies database connectivity
func Ping() error {
	if instance == nil {
		if dbErr != nil {
			return dbErr
		}
		return ErrNotInitialized
	}

	sqlDB, err := instance.DB()
	if err != nil {
		return err
	}

	return sqlDB.Ping()
}

func ResetForTesting() {
	instance = nil
	dbErr = nil
	once = sync.Once{}
}
