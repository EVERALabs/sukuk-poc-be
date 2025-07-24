package models

// This file contains common model utilities and exports

// AllModels returns a slice of all models for migrations
func AllModels() []interface{} {
	return []interface{}{
		&Company{},
		&Sukuk{},        // Renamed from SukukSeries
		&Investment{},
		&Yield{},        // Renamed from YieldClaim
		&Redemption{},
		&SystemState{},
		&SukukMetadata{}, // New model for onchain + offchain metadata
		&SukukPurchased{}, // Blockchain event for sukuk purchases
		&RedemptionRequested{}, // Blockchain event for redemption requests
		// Removed &Event{} as we'll use indexer's blockchain.events table
	}
}