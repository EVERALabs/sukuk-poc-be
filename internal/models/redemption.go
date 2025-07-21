package models

import (
	"time"

	"gorm.io/gorm"
)

// RedemptionStatus represents the status of a redemption request
type RedemptionStatus string

const (
	RedemptionStatusRequested RedemptionStatus = "requested" // Redemption requested but not processed
	RedemptionStatusCompleted RedemptionStatus = "completed" // Redemption completed and IDRX transferred
	RedemptionStatusCancelled RedemptionStatus = "cancelled" // Redemption cancelled
)

// Redemption represents a redemption request for sukuk tokens
type Redemption struct {
	ID               uint             `gorm:"primaryKey" json:"id"`
	SukukID          uint             `gorm:"not null;index;column:sukuk_series_id" json:"sukuk_id"` // FK to sukuk_series table
	InvestorAddress  string           `gorm:"size:42;not null;index" json:"investor_address"`
	TokenAmount      string           `gorm:"size:78;not null" json:"token_amount"`      // Sukuk tokens to redeem in wei
	RedemptionAmount string           `gorm:"size:78;not null" json:"redemption_amount"` // IDRX amount to receive in wei
	RequestTxHash    string           `gorm:"size:66;not null;index" json:"request_tx_hash"`
	RequestLogIndex  int              `gorm:"not null" json:"request_log_index"`
	CompleteTxHash   string           `gorm:"size:66;index" json:"complete_tx_hash,omitempty"`
	CompleteLogIndex *int             `json:"complete_log_index,omitempty"`
	RequestDate      time.Time        `gorm:"not null" json:"request_date"`
	CompletedAt      *time.Time       `json:"completed_at,omitempty"`
	Status           RedemptionStatus `gorm:"size:20;not null;default:'requested'" json:"status"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
	DeletedAt        gorm.DeletedAt   `gorm:"index" json:"-"`

	// Relationships
	Sukuk Sukuk `gorm:"foreignKey:SukukID;references:ID" json:"sukuk,omitempty"`
}

// TableName returns the table name for Redemption model
func (Redemption) TableName() string {
	return "redemptions"
}

// BeforeCreate hook to normalize investor address and set defaults
func (r *Redemption) BeforeCreate(tx *gorm.DB) error {
	r.InvestorAddress = normalizeAddress(r.InvestorAddress)
	if r.Status == "" {
		r.Status = RedemptionStatusRequested
	}
	return nil
}

// BeforeUpdate hook to normalize investor address
func (r *Redemption) BeforeUpdate(tx *gorm.DB) error {
	if r.InvestorAddress != "" {
		r.InvestorAddress = normalizeAddress(r.InvestorAddress)
	}
	return nil
}

// IsPending returns true if the redemption is pending
func (r *Redemption) IsPending() bool {
	return r.Status == RedemptionStatusRequested
}

// IsCompleted returns true if the redemption is completed
func (r *Redemption) IsCompleted() bool {
	return r.Status == RedemptionStatusCompleted
}

// IsCancelled returns true if the redemption is cancelled
func (r *Redemption) IsCancelled() bool {
	return r.Status == RedemptionStatusCancelled
}

// MarkCompleted marks the redemption as completed
func (r *Redemption) MarkCompleted(db *gorm.DB, txHash string, logIndex int) error {
	now := time.Now()
	r.Status = RedemptionStatusCompleted
	r.CompleteTxHash = txHash
	r.CompleteLogIndex = &logIndex
	r.CompletedAt = &now
	return db.Save(r).Error
}

// GetProcessingTime returns the time taken to process the redemption
func (r *Redemption) GetProcessingTime() *time.Duration {
	if r.CompletedAt == nil {
		return nil
	}
	duration := r.CompletedAt.Sub(r.RequestDate)
	return &duration
}

// GetInvestorRedemptions returns all redemptions for a specific investor and sukuk
func GetInvestorRedemptions(db *gorm.DB, sukukID uint, investorAddress string) ([]Redemption, error) {
	var redemptions []Redemption
	normalizedAddress := normalizeAddress(investorAddress)
	err := db.Where("sukuk_id = ? AND investor_address = ?", sukukID, normalizedAddress).
		Order("created_at DESC").
		Find(&redemptions).Error
	return redemptions, err
}

// GetSukukRedemptions returns all redemptions for a specific sukuk
func GetSukukRedemptions(db *gorm.DB, sukukID uint) ([]Redemption, error) {
	var redemptions []Redemption
	err := db.Preload("Sukuk").Where("sukuk_id = ?", sukukID).
		Order("created_at DESC").
		Find(&redemptions).Error
	return redemptions, err
}

// GetPendingRedemptions returns all pending redemptions across all sukuks
func GetPendingRedemptions(db *gorm.DB) ([]Redemption, error) {
	var redemptions []Redemption
	err := db.Preload("Sukuk").Preload("Sukuk.Company").
		Where("status = ?", RedemptionStatusRequested).
		Order("created_at ASC"). // Oldest first for processing
		Find(&redemptions).Error
	return redemptions, err
}

// GetRedemptionsByDateRange returns redemptions within a date range
func GetRedemptionsByDateRange(db *gorm.DB, startDate, endDate time.Time) ([]Redemption, error) {
	var redemptions []Redemption
	err := db.Preload("Sukuk").Preload("Sukuk.Company").
		Where("request_date BETWEEN ? AND ?", startDate, endDate).
		Order("request_date DESC").
		Find(&redemptions).Error
	return redemptions, err
}