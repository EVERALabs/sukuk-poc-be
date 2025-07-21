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
		// Removed &Event{} as we'll use indexer's blockchain.events table
	}
}