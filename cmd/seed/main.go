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

	// Seed database with sample data
	db := database.GetDB()
	if err := database.SeedData(db); err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	log.Println("Database seeding completed successfully!")
}
