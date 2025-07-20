package database

import (
	"fmt"

	"github.com/kadzu/sukuk-poc-be/internal/config"
	"github.com/kadzu/sukuk-poc-be/internal/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// CreateDatabaseIfNotExists creates the database if it doesn't exist
func CreateDatabaseIfNotExists(cfg *config.Config) error {
	// Connect to postgres database (default) to create our target database
	adminDSN := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=postgres port=%d sslmode=%s TimeZone=UTC",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)

	// Use silent logger for admin connection
	adminDB, err := gorm.Open(postgres.Open(adminDSN), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to postgres database: %w", err)
	}

	// Get underlying sql.DB
	sqlDB, err := adminDB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	defer sqlDB.Close()

	// Check if database exists
	var exists bool
	err = adminDB.Raw("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = ?)", cfg.Database.DBName).Scan(&exists).Error
	if err != nil {
		return fmt.Errorf("failed to check if database exists: %w", err)
	}

	if !exists {
		logger.WithField("database", cfg.Database.DBName).Info("Database does not exist, creating it...")
		
		// Create database
		createSQL := fmt.Sprintf("CREATE DATABASE %s", cfg.Database.DBName)
		if err := adminDB.Exec(createSQL).Error; err != nil {
			logger.WithError(err).WithField("database", cfg.Database.DBName).Error("Failed to create database")
			return fmt.Errorf("failed to create database '%s': %w", cfg.Database.DBName, err)
		}
		
		logger.WithField("database", cfg.Database.DBName).Info("Database created successfully")
	} else {
		logger.WithField("database", cfg.Database.DBName).Info("Database already exists")
	}

	return nil
}

// SetupDatabase creates database if needed, connects, and runs migrations
func SetupDatabase(cfg *config.Config) error {
	// Step 1: Create database if it doesn't exist
	if err := CreateDatabaseIfNotExists(cfg); err != nil {
		return fmt.Errorf("database setup failed: %w", err)
	}

	// Step 2: Connect to the target database
	if err := Connect(cfg); err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	// Step 3: Run migrations
	if err := Migrate(); err != nil {
		return fmt.Errorf("database migration failed: %w", err)
	}

	return nil
}