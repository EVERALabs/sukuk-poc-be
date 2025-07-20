package models

// This file contains common model utilities and exports

// AllModels returns a slice of all models for migrations
func AllModels() []interface{} {
	return []interface{}{
		&Company{},
		&SukukSeries{},
		&Investment{},
		&YieldClaim{},
		&Redemption{},
		&Event{},
	}
}