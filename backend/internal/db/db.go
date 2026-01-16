package db

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/rootlock/rootlock/backend/config"
)

// DB holds the database connection
type DB struct {
	*gorm.DB
}

// New creates a new database connection
func New(cfg *config.Config) (*DB, error) {
	// Configure GORM logger
	logLevel := logger.Silent
	if cfg.IsDevelopment() {
		logLevel = logger.Info
	}

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	// Open database connection
	db, err := gorm.Open(postgres.Open(cfg.Database.URL), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB to set connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{DB: db}, nil
}

// Close closes the database connection
func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// AutoMigrate runs auto-migration for all models
// Note: This is for development only. Use migrations in production.
func (db *DB) AutoMigrate() error {
	return db.DB.AutoMigrate(
		&User{},
		&Vault{},
		&Device{},
		&Passkey{},
		&AuditLog{},
		&RateLimit{},
		&WebAuthnChallenge{},
		&Session{},
	)
}

// Health checks if the database connection is healthy
func (db *DB) Health() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
