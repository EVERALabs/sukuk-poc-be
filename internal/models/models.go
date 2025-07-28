package models

// This file contains common model utilities and exports

// AllModels returns a slice of all models for migrations
func AllModels() []interface{} {
	return []interface{}{
		&SystemState{},
		&SukukMetadata{}, // Model for onchain + offchain metadata
		&SukukPurchased{}, // Blockchain event for sukuk purchases
		&RedemptionRequested{}, // Blockchain event for redemption requests
		// Only keeping essential models for indexer data + metadata
	}
}