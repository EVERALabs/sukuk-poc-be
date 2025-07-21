package models

import (
	"time"

	"gorm.io/gorm"
)

// YieldStatus represents the status of a yield claim
type YieldStatus string

const (
	YieldStatusPending YieldStatus = "pending" // Yield distributed but not claimed
	YieldStatusClaimed YieldStatus = "claimed" // Yield claimed by investor
	YieldStatusExpired YieldStatus = "expired" // Yield claim expired (if applicable)
)

// Yield represents a yield distribution and claim (renamed from YieldClaim)
type Yield struct {
	ID               uint        `gorm:"primaryKey" json:"id"`
	SukukID          uint        `gorm:"not null;index;column:sukuk_series_id" json:"sukuk_id"` // FK to sukuk_series table
	InvestorAddress  string      `gorm:"size:42;not null;index" json:"investor_address"`
	YieldAmount      string      `gorm:"size:78;not null" json:"yield_amount"`        // IDRX yield amount in wei
	DistributionDate time.Time   `gorm:"not null;index" json:"distribution_date"`    // When yield was distributed
	ClaimDate        *time.Time  `json:"claim_date,omitempty"`                       // When yield was claimed
	DistTxHash       string      `gorm:"size:66;not null;index" json:"dist_tx_hash"` // Distribution transaction hash
	DistLogIndex     int         `gorm:"not null" json:"dist_log_index"`             // Distribution log index
	ClaimTxHash      string      `gorm:"size:66;index" json:"claim_tx_hash,omitempty"` // Claim transaction hash
	ClaimLogIndex    *int        `json:"claim_log_index,omitempty"`                  // Claim log index
	PeriodStart      time.Time   `gorm:"not null" json:"period_start"`               // Yield period start
	PeriodEnd        time.Time   `gorm:"not null" json:"period_end"`                 // Yield period end
	Status           YieldStatus `gorm:"size:20;not null;default:'pending'" json:"status"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Sukuk Sukuk `gorm:"foreignKey:SukukID;references:ID" json:"sukuk,omitempty"`
}

// TableName returns the table name for Yield model
func (Yield) TableName() string {
	return "yield_claims" // Keep existing table name for backward compatibility
}

// BeforeCreate hook to normalize investor address and set defaults
func (y *Yield) BeforeCreate(tx *gorm.DB) error {
	y.InvestorAddress = normalizeAddress(y.InvestorAddress)
	if y.Status == "" {
		y.Status = YieldStatusPending
	}
	return nil
}

// BeforeUpdate hook to normalize investor address
func (y *Yield) BeforeUpdate(tx *gorm.DB) error {
	if y.InvestorAddress != "" {
		y.InvestorAddress = normalizeAddress(y.InvestorAddress)
	}
	return nil
}

// IsPending returns true if the yield is pending
func (y *Yield) IsPending() bool {
	return y.Status == YieldStatusPending
}

// IsClaimed returns true if the yield has been claimed
func (y *Yield) IsClaimed() bool {
	return y.Status == YieldStatusClaimed
}

// IsExpired returns true if the yield claim has expired
func (y *Yield) IsExpired() bool {
	return y.Status == YieldStatusExpired
}

// MarkClaimed marks the yield as claimed
func (y *Yield) MarkClaimed(db *gorm.DB, txHash string, logIndex int) error {
	now := time.Now()
	y.Status = YieldStatusClaimed
	y.ClaimTxHash = txHash
	y.ClaimLogIndex = &logIndex
	y.ClaimDate = &now
	return db.Save(y).Error
}

// GetClaimWaitTime returns the time between distribution and claim
func (y *Yield) GetClaimWaitTime() *time.Duration {
	if y.ClaimDate == nil {
		return nil
	}
	duration := y.ClaimDate.Sub(y.DistributionDate)
	return &duration
}

// GetYieldPeriodDuration returns the duration of the yield period
func (y *Yield) GetYieldPeriodDuration() time.Duration {
	return y.PeriodEnd.Sub(y.PeriodStart)
}

// GetInvestorYields returns all yields for a specific investor and sukuk
func GetInvestorYields(db *gorm.DB, sukukID uint, investorAddress string) ([]Yield, error) {
	var yields []Yield
	normalizedAddress := normalizeAddress(investorAddress)
	err := db.Where("sukuk_id = ? AND investor_address = ?", sukukID, normalizedAddress).
		Order("distribution_date DESC").
		Find(&yields).Error
	return yields, err
}

// GetSukukYields returns all yields for a specific sukuk
func GetSukukYields(db *gorm.DB, sukukID uint) ([]Yield, error) {
	var yields []Yield
	err := db.Preload("Sukuk").Where("sukuk_id = ?", sukukID).
		Order("distribution_date DESC").
		Find(&yields).Error
	return yields, err
}

// GetPendingYields returns all pending yields across all sukuks
func GetPendingYields(db *gorm.DB) ([]Yield, error) {
	var yields []Yield
	err := db.Preload("Sukuk").Preload("Sukuk.Company").
		Where("status = ?", YieldStatusPending).
		Order("distribution_date ASC"). // Oldest distributions first
		Find(&yields).Error
	return yields, err
}

// GetYieldsByDistributionDate returns all yields for a specific distribution date
func GetYieldsByDistributionDate(db *gorm.DB, distributionDate time.Time) ([]Yield, error) {
	var yields []Yield
	startOfDay := time.Date(distributionDate.Year(), distributionDate.Month(), distributionDate.Day(), 0, 0, 0, 0, distributionDate.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)
	
	err := db.Preload("Sukuk").Preload("Sukuk.Company").
		Where("distribution_date >= ? AND distribution_date < ?", startOfDay, endOfDay).
		Order("distribution_date ASC").
		Find(&yields).Error
	return yields, err
}

// GetYieldsByDateRange returns yields within a date range
func GetYieldsByDateRange(db *gorm.DB, startDate, endDate time.Time) ([]Yield, error) {
	var yields []Yield
	err := db.Preload("Sukuk").Preload("Sukuk.Company").
		Where("distribution_date BETWEEN ? AND ?", startDate, endDate).
		Order("distribution_date DESC").
		Find(&yields).Error
	return yields, err
}

// GetYieldDistributionSummary returns summary of yield distributions grouped by date and sukuk
func GetYieldDistributionSummary(db *gorm.DB) ([]map[string]interface{}, error) {
	var results []map[string]interface{}
	
	err := db.Model(&Yield{}).
		Select(`
			DATE(distribution_date) as date,
			sukuk_id,
			COUNT(*) as total_claims,
			SUM(CASE WHEN status = 'claimed' THEN 1 ELSE 0 END) as claimed_count,
			SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending_count,
			SUM(CAST(yield_amount AS NUMERIC)) as total_amount
		`).
		Group("DATE(distribution_date), sukuk_id").
		Order("date DESC").
		Find(&results).Error
	
	return results, err
}