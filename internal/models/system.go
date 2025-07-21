package models

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// SystemState represents system configuration and state tracking
type SystemState struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Key       string    `gorm:"size:255;uniqueIndex;not null" json:"key"`
	Value     string    `gorm:"type:text;not null" json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName returns the table name for SystemState model
func (SystemState) TableName() string {
	return "system_states"
}

// System state keys
const (
	LastProcessedEventIDKey = "last_processed_event_id"
	SyncStatusKey          = "sync_status"
	LastSyncTimeKey        = "last_sync_time"
	BlockchainHeightKey    = "blockchain_height"
	IndexerVersionKey      = "indexer_version"
	MaintenanceModeKey     = "maintenance_mode"
)

// System state values
const (
	SyncStatusActive    = "active"
	SyncStatusPaused    = "paused"
	SyncStatusError     = "error"
	MaintenanceModeOn   = "on"
	MaintenanceModeOff  = "off"
)

// GetSystemState retrieves a system state by key
func GetSystemState(db *gorm.DB, key string) (*SystemState, error) {
	var state SystemState
	err := db.Where("key = ?", key).First(&state).Error
	if err != nil {
		return nil, err
	}
	return &state, nil
}

// SetSystemState sets or updates a system state
func SetSystemState(db *gorm.DB, key, value string) error {
	var state SystemState
	err := db.Where("key = ?", key).First(&state).Error
	
	if err == gorm.ErrRecordNotFound {
		// Create new record
		state = SystemState{
			Key:   key,
			Value: value,
		}
		return db.Create(&state).Error
	} else if err != nil {
		return err
	}
	
	// Update existing record
	state.Value = value
	return db.Save(&state).Error
}

// GetLastProcessedEventID returns the last processed blockchain event ID
func GetLastProcessedEventID(db *gorm.DB) (string, error) {
	state, err := GetSystemState(db, LastProcessedEventIDKey)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "0", nil // Default to 0 if not set
		}
		return "", err
	}
	return state.Value, nil
}

// SetLastProcessedEventID sets the last processed blockchain event ID
func SetLastProcessedEventID(db *gorm.DB, eventID string) error {
	return SetSystemState(db, LastProcessedEventIDKey, eventID)
}

// GetSyncStatus returns the current sync status
func GetSyncStatus(db *gorm.DB) (string, error) {
	state, err := GetSystemState(db, SyncStatusKey)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return SyncStatusActive, nil // Default to active
		}
		return "", err
	}
	return state.Value, nil
}

// SetSyncStatus sets the sync status
func SetSyncStatus(db *gorm.DB, status string) error {
	return SetSystemState(db, SyncStatusKey, status)
}

// GetLastSyncTime returns the last sync time
func GetLastSyncTime(db *gorm.DB) (time.Time, error) {
	state, err := GetSystemState(db, LastSyncTimeKey)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return time.Time{}, nil // Return zero time if not set
		}
		return time.Time{}, err
	}
	
	return time.Parse(time.RFC3339, state.Value)
}

// SetLastSyncTime sets the last sync time
func SetLastSyncTime(db *gorm.DB, syncTime time.Time) error {
	return SetSystemState(db, LastSyncTimeKey, syncTime.Format(time.RFC3339))
}

// IsMaintenanceMode checks if the system is in maintenance mode
func IsMaintenanceMode(db *gorm.DB) (bool, error) {
	state, err := GetSystemState(db, MaintenanceModeKey)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil // Default to off
		}
		return false, err
	}
	return state.Value == MaintenanceModeOn, nil
}

// SetMaintenanceMode sets the maintenance mode
func SetMaintenanceMode(db *gorm.DB, enabled bool) error {
	value := MaintenanceModeOff
	if enabled {
		value = MaintenanceModeOn
	}
	return SetSystemState(db, MaintenanceModeKey, value)
}

// GetBlockchainHeight returns the current blockchain height
func GetBlockchainHeight(db *gorm.DB) (string, error) {
	state, err := GetSystemState(db, BlockchainHeightKey)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "0", nil // Default to 0
		}
		return "", err
	}
	return state.Value, nil
}

// SetBlockchainHeight sets the current blockchain height
func SetBlockchainHeight(db *gorm.DB, height string) error {
	return SetSystemState(db, BlockchainHeightKey, height)
}

// GetIndexerVersion returns the indexer version
func GetIndexerVersion(db *gorm.DB) (string, error) {
	state, err := GetSystemState(db, IndexerVersionKey)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "unknown", nil
		}
		return "", err
	}
	return state.Value, nil
}

// SetIndexerVersion sets the indexer version
func SetIndexerVersion(db *gorm.DB, version string) error {
	return SetSystemState(db, IndexerVersionKey, version)
}

// GetAllSystemStates returns all system states
func GetAllSystemStates(db *gorm.DB) ([]SystemState, error) {
	var states []SystemState
	err := db.Order("key ASC").Find(&states).Error
	return states, err
}

// normalizeAddress converts an Ethereum address to lowercase
func normalizeAddress(address string) string {
	return strings.ToLower(address)
}