package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// JSON is a custom type for storing JSON data
type JSON map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSON) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, j)
	case string:
		return json.Unmarshal([]byte(v), j)
	default:
		return fmt.Errorf("cannot scan %T into JSON", value)
	}
}

// Event represents a blockchain event processed by the indexer
type Event struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	EventName       string         `gorm:"index;not null;size:100" json:"event_name"`
	BlockNumber     int64          `gorm:"index;not null" json:"block_number"`
	TxHash          string         `gorm:"index;not null;size:66" json:"tx_hash"`
	ContractAddress string         `gorm:"index;not null;size:42" json:"contract_address"`
	Data            JSON           `gorm:"type:jsonb" json:"data"`
	Processed       bool           `gorm:"index;default:false" json:"processed"`
	ProcessedAt     *time.Time     `json:"processed_at,omitempty"`
	ChainID         int64          `gorm:"default:84532" json:"chain_id"` // Base Testnet
	Error           string         `gorm:"size:500" json:"error,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// BeforeCreate hook to normalize addresses and hash
func (e *Event) BeforeCreate(tx *gorm.DB) error {
	e.TxHash = strings.ToLower(e.TxHash)
	e.ContractAddress = strings.ToLower(e.ContractAddress)
	return nil
}

// MarkAsProcessed marks the event as processed
func (e *Event) MarkAsProcessed(tx *gorm.DB) error {
	now := time.Now()
	e.Processed = true
	e.ProcessedAt = &now
	return tx.Save(e).Error
}

// MarkAsError marks the event as failed with error message
func (e *Event) MarkAsError(tx *gorm.DB, errorMsg string) error {
	e.Error = errorMsg
	return tx.Save(e).Error
}

// TableName specifies the table name
func (Event) TableName() string {
	return "events"
}