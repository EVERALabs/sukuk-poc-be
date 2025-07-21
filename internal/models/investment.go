package models

import (
	"time"

	"gorm.io/gorm"
)

// InvestmentStatus represents the status of an investment
type InvestmentStatus string

const (
	InvestmentStatusActive   InvestmentStatus = "active"   // Investment is active
	InvestmentStatusRedeemed InvestmentStatus = "redeemed" // Fully redeemed
	InvestmentStatusPartial  InvestmentStatus = "partial"  // Partially redeemed
)

// Investment represents an investor's investment in a sukuk
type Investment struct {
	ID               uint             `gorm:"primaryKey" json:"id"`
	SukukID          uint             `gorm:"not null;index;column:sukuk_series_id" json:"sukuk_id"` // FK to sukuk_series table
	InvestorAddress  string           `gorm:"size:42;not null;index" json:"investor_address"`
	InvestmentAmount string           `gorm:"size:78;not null" json:"investment_amount"` // IDRX amount in wei
	TokenAmount      string           `gorm:"size:78;not null" json:"token_amount"`      // Sukuk tokens received in wei
	TokenPrice       string           `gorm:"size:78;not null" json:"token_price"`       // Price per token in wei
	TxHash           string           `gorm:"size:66;not null;index" json:"tx_hash"`
	LogIndex         int              `gorm:"not null" json:"log_index"`
	InvestmentDate   time.Time        `gorm:"not null" json:"investment_date"`
	Status           InvestmentStatus `gorm:"size:20;not null;default:'active'" json:"status"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
	DeletedAt        gorm.DeletedAt   `gorm:"index" json:"-"`

	// Relationships
	Sukuk Sukuk `gorm:"foreignKey:SukukID;references:ID" json:"sukuk,omitempty"`
}

// TableName returns the table name for Investment model
func (Investment) TableName() string {
	return "investments"
}

// BeforeCreate hook to normalize investor address
func (i *Investment) BeforeCreate(tx *gorm.DB) error {
	i.InvestorAddress = normalizeAddress(i.InvestorAddress)
	if i.Status == "" {
		i.Status = InvestmentStatusActive
	}
	return nil
}

// BeforeUpdate hook to normalize investor address
func (i *Investment) BeforeUpdate(tx *gorm.DB) error {
	if i.InvestorAddress != "" {
		i.InvestorAddress = normalizeAddress(i.InvestorAddress)
	}
	return nil
}

// IsActive returns true if the investment is active
func (i *Investment) IsActive() bool {
	return i.Status == InvestmentStatusActive
}

// IsFullyRedeemed returns true if the investment is fully redeemed
func (i *Investment) IsFullyRedeemed() bool {
	return i.Status == InvestmentStatusRedeemed
}

// GetYields returns all yield claims for this investment
func (i *Investment) GetYields(db *gorm.DB) ([]Yield, error) {
	var yields []Yield
	err := db.Where("sukuk_id = ? AND investor_address = ?", i.SukukID, i.InvestorAddress).Find(&yields).Error
	return yields, err
}

// GetRedemptions returns all redemptions for this investment
func (i *Investment) GetRedemptions(db *gorm.DB) ([]Redemption, error) {
	var redemptions []Redemption
	err := db.Where("sukuk_id = ? AND investor_address = ?", i.SukukID, i.InvestorAddress).Find(&redemptions).Error
	return redemptions, err
}

// GetTotalYieldsClaimed returns the total yield amount claimed for this investment
func (i *Investment) GetTotalYieldsClaimed(db *gorm.DB) (string, error) {
	var result struct {
		Total string
	}
	
	err := db.Model(&Yield{}).
		Select("COALESCE(SUM(CAST(yield_amount AS NUMERIC)), 0) as total").
		Where("sukuk_id = ? AND investor_address = ? AND status = ?", 
			i.SukukID, i.InvestorAddress, YieldStatusClaimed).
		Scan(&result).Error
	
	if err != nil {
		return "0", err
	}
	
	if result.Total == "" {
		return "0", nil
	}
	
	return result.Total, nil
}

// GetTotalRedemptionAmount returns the total amount redeemed for this investment
func (i *Investment) GetTotalRedemptionAmount(db *gorm.DB) (string, error) {
	var result struct {
		Total string
	}
	
	err := db.Model(&Redemption{}).
		Select("COALESCE(SUM(CAST(redemption_amount AS NUMERIC)), 0) as total").
		Where("sukuk_id = ? AND investor_address = ? AND status = ?", 
			i.SukukID, i.InvestorAddress, RedemptionStatusCompleted).
		Scan(&result).Error
	
	if err != nil {
		return "0", err
	}
	
	if result.Total == "" {
		return "0", nil
	}
	
	return result.Total, nil
}