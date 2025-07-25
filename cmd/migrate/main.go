package main

import (
	"log"

	"sukuk-be/internal/config"
	"sukuk-be/internal/database"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup database (create if needed, connect, migrate)
	if err := database.SetupDatabase(cfg); err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}
	defer database.Close()

	log.Println("Migrations completed successfully!")
}
