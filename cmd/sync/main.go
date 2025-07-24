package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sukuk-be/internal/config"
	"sukuk-be/internal/database"
	"sukuk-be/internal/services"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	database.Init(cfg)

	// Create blockchain event sync service
	eventSync := services.NewBlockchainEventSync(cfg)

	// Create context that can be cancelled
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Starting blockchain event auto-sync service...")

	// Start auto-sync in a goroutine
	go func() {
		// Sync every 30 seconds
		eventSync.StartAutoSync(ctx, 30*time.Second)
	}()

	// Wait for signal
	<-sigChan
	log.Println("Received shutdown signal, stopping sync service...")

	// Cancel context to stop sync
	cancel()

	// Give some time for graceful shutdown
	time.Sleep(2 * time.Second)
	log.Println("Auto-sync service stopped")
}