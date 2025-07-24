package models

import (
	"time"

	"gorm.io/gorm"
)

// SukukPurchased represents a sukuk purchase event from the blockchain
type SukukPurchased struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Buyer         string         `gorm:"size:42;not null;index" json:"buyer"`
	SukukAddress  string         `gorm:"size:42;not null;index" json:"sukuk_address"`
	PaymentToken  string         `gorm:"size:42;not null" json:"payment_token"`
	Amount        string         `gorm:"size:78;not null" json:"amount"`
	BlockNumber   uint64         `gorm:"not null;index" json:"block_number"`
	TxHash        string         `gorm:"size:66;not null;index" json:"tx_hash"`
	LogIndex      uint           `gorm:"not null" json:"log_index"`
	Timestamp     time.Time      `gorm:"not null;index" json:"timestamp"`
	Processed     bool           `gorm:"default:false;index" json:"processed"`
	ProcessedAt   *time.Time     `json:"processed_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the table name for SukukPurchased model
func (SukukPurchased) TableName() string {
	return "sukuk_purchased_events"
}

// BeforeCreate hook to normalize addresses
func (sp *SukukPurchased) BeforeCreate(tx *gorm.DB) error {
	sp.Buyer = normalizeAddress(sp.Buyer)
	sp.SukukAddress = normalizeAddress(sp.SukukAddress)
	sp.PaymentToken = normalizeAddress(sp.PaymentToken)
	return nil
}

// RedemptionRequested represents a redemption request event from the blockchain
type RedemptionRequested struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	User          string         `gorm:"size:42;not null;index" json:"user"`
	SukukAddress  string         `gorm:"size:42;not null;index" json:"sukuk_address"`
	Amount        string         `gorm:"size:78;not null" json:"amount"`
	PaymentToken  string         `gorm:"size:42;not null" json:"payment_token"`
	BlockNumber   uint64         `gorm:"not null;index" json:"block_number"`
	TxHash        string         `gorm:"size:66;not null;index" json:"tx_hash"`
	LogIndex      uint           `gorm:"not null" json:"log_index"`
	Timestamp     time.Time      `gorm:"not null;index" json:"timestamp"`
	Processed     bool           `gorm:"default:false;index" json:"processed"`
	ProcessedAt   *time.Time     `json:"processed_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName returns the table name for RedemptionRequested model
func (RedemptionRequested) TableName() string {
	return "redemption_requested_events"
}

// BeforeCreate hook to normalize addresses
func (rr *RedemptionRequested) BeforeCreate(tx *gorm.DB) error {
	rr.User = normalizeAddress(rr.User)
	rr.SukukAddress = normalizeAddress(rr.SukukAddress)
	rr.PaymentToken = normalizeAddress(rr.PaymentToken)
	return nil
}


// GetUnprocessedSukukPurchases retrieves unprocessed sukuk purchase events
func GetUnprocessedSukukPurchases(db *gorm.DB, limit int) ([]SukukPurchased, error) {
	var events []SukukPurchased
	query := db.Where("processed = ?", false).
		Order("block_number ASC, log_index ASC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&events).Error
	return events, err
}

// GetUnprocessedRedemptionRequests retrieves unprocessed redemption request events
func GetUnprocessedRedemptionRequests(db *gorm.DB, limit int) ([]RedemptionRequested, error) {
	var events []RedemptionRequested
	query := db.Where("processed = ?", false).
		Order("block_number ASC, log_index ASC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&events).Error
	return events, err
}

// MarkAsProcessed marks the event as processed
func (sp *SukukPurchased) MarkAsProcessed(db *gorm.DB) error {
	now := time.Now()
	return db.Model(sp).Updates(map[string]interface{}{
		"processed":    true,
		"processed_at": &now,
	}).Error
}

// MarkAsProcessed marks the event as processed
func (rr *RedemptionRequested) MarkAsProcessed(db *gorm.DB) error {
	now := time.Now()
	return db.Model(rr).Updates(map[string]interface{}{
		"processed":    true,
		"processed_at": &now,
	}).Error
}

// GetSukukPurchaseByTxHashAndLogIndex retrieves a sukuk purchase event by transaction hash and log index
func GetSukukPurchaseByTxHashAndLogIndex(db *gorm.DB, txHash string, logIndex uint) (*SukukPurchased, error) {
	var event SukukPurchased
	err := db.Where("tx_hash = ? AND log_index = ?", txHash, logIndex).First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

// GetRedemptionRequestByTxHashAndLogIndex retrieves a redemption request event by transaction hash and log index
func GetRedemptionRequestByTxHashAndLogIndex(db *gorm.DB, txHash string, logIndex uint) (*RedemptionRequested, error) {
	var event RedemptionRequested
	err := db.Where("tx_hash = ? AND log_index = ?", txHash, logIndex).First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}