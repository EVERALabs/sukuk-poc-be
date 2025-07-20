package main

import (
	"github.com/kadzu/sukuk-poc-be/internal/config"
	"github.com/kadzu/sukuk-poc-be/internal/database"
	"github.com/kadzu/sukuk-poc-be/internal/logger"
	"github.com/kadzu/sukuk-poc-be/internal/server"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger with configuration
	logger.Init(cfg.Logger.Level, cfg.Logger.Format)
	logger.Info("Starting Sukuk POC API server")

	// Setup database (create if needed, connect, migrate)
	if err := database.SetupDatabase(cfg); err != nil {
		logger.Fatalf("Failed to setup database: %v", err)
	}
	defer database.Close()

	// Start server
	srv := server.New(cfg)
	logger.WithField("port", cfg.App.Port).Info("Server starting")
	
	if err := srv.Start(); err != nil {
		logger.Fatalf("Server failed to start: %v", err)
	}
}