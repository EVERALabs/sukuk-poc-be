package database

import (
	"context"
	"fmt"
	"time"

	"sukuk-be/internal/config"
	"sukuk-be/internal/logger"
	"sukuk-be/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect establishes database connection
func Connect(cfg *config.Config) error {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=UTC",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)

	// Configure GORM logger based on environment
	var gormLogLevel gormLogger.LogLevel
	if cfg.App.Debug {
		gormLogLevel = gormLogger.Info
	} else {
		gormLogLevel = gormLogger.Silent
	}

	gormLog := gormLogger.Default.LogMode(gormLogLevel)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLog,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	// Test the connection
	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	DB = db
	logger.WithFields(map[string]interface{}{
		"host":              cfg.Database.Host,
		"port":              cfg.Database.Port,
		"database":          cfg.Database.DBName,
		"max_open_conns":    cfg.Database.MaxOpenConns,
		"max_idle_conns":    cfg.Database.MaxIdleConns,
		"conn_max_lifetime": cfg.Database.ConnMaxLifetime.String(),
	}).Info("Database connection established successfully")
	return nil
}

// Migrate runs database migrations
func Migrate() error {
	if DB == nil {
		return fmt.Errorf("database connection not established")
	}

	logger.Info("Running database migrations...")

	// Auto-migrate all models
	if err := DB.AutoMigrate(models.AllModels()...); err != nil {
		logger.WithError(err).Error("Failed to run database migrations")
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	logger.Info("Database migrations completed successfully")
	return nil
}

// Close closes the database connection
func Close() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	return sqlDB.Close()
}

// GetDB returns the database instance
func GetDB() *gorm.DB {
	return DB
}

// Health checks database connection health
func Health() error {
	if DB == nil {
		return fmt.Errorf("database connection not established")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return sqlDB.PingContext(ctx)
}
