package models

import (
	"time"

	"gorm.io/gorm"
)

// SukukStatus represents the status of a sukuk series
type SukukStatus string

const (
	SukukStatusDraft   SukukStatus = "draft"   // Created but not deployed
	SukukStatusActive  SukukStatus = "active"  // Deployed and accepting investments
	SukukStatusPaused  SukukStatus = "paused"  // Temporarily not accepting investments
	SukukStatusMatured SukukStatus = "matured" // Reached maturity date
	SukukStatusClosed  SukukStatus = "closed"  // Fully redeemed and closed
)

// Sukuk represents a sukuk series (renamed from SukukSeries)
type Sukuk struct {
	ID                uint        `gorm:"primaryKey" json:"id"`
	CompanyID         uint        `gorm:"not null;index" json:"company_id"`
	Name              string      `gorm:"size:255;not null" json:"name"`
	Symbol            string      `gorm:"size:20;not null" json:"symbol"`
	Description       string      `gorm:"type:text" json:"description"`
	TotalSupply       string      `gorm:"size:78;not null" json:"total_supply"`         // Wei string
	OutstandingSupply string      `gorm:"size:78;not null;default:'0'" json:"outstanding_supply"` // Wei string
	YieldRate         float64     `gorm:"type:decimal(5,4);not null" json:"yield_rate"` // e.g., 0.085 for 8.5%
	MaturityDate      time.Time   `gorm:"not null" json:"maturity_date"`
	PaymentFrequency  int         `gorm:"not null;default:4" json:"payment_frequency"` // Times per year
	MinInvestment     string      `gorm:"size:78;not null" json:"min_investment"`      // Wei string
	MaxInvestment     string      `gorm:"size:78" json:"max_investment"`               // Wei string, nullable
	TokenAddress      string      `gorm:"size:42;uniqueIndex" json:"token_address"`    // Ethereum contract address
	Status            SukukStatus `gorm:"size:20;not null;default:'draft'" json:"status"`
	IsRedeemable      bool        `gorm:"default:true" json:"is_redeemable"`
	Prospectus        string      `gorm:"size:500" json:"prospectus"` // File path
	CreatedAt         time.Time   `json:"created_at"`
	UpdatedAt         time.Time   `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Company     Company      `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Investments []Investment `gorm:"foreignKey:SukukID" json:"investments,omitempty"`
	Yields      []Yield      `gorm:"foreignKey:SukukID" json:"yields,omitempty"`
	Redemptions []Redemption `gorm:"foreignKey:SukukID" json:"redemptions,omitempty"`
}

// TableName returns the table name for Sukuk model
func (Sukuk) TableName() string {
	return "sukuk_series" // Keep existing table name for backward compatibility
}

// BeforeCreate hook to set defaults
func (s *Sukuk) BeforeCreate(tx *gorm.DB) error {
	if s.Status == "" {
		s.Status = SukukStatusDraft
	}
	if s.OutstandingSupply == "" {
		s.OutstandingSupply = "0"
	}
	if s.PaymentFrequency == 0 {
		s.PaymentFrequency = 4 // Quarterly by default
	}
	return nil
}

// BeforeUpdate hook to normalize token address
func (s *Sukuk) BeforeUpdate(tx *gorm.DB) error {
	if s.TokenAddress != "" {
		s.TokenAddress = normalizeAddress(s.TokenAddress)
	}
	return nil
}

// IsActive returns true if the sukuk is in active status
func (s *Sukuk) IsActive() bool {
	return s.Status == SukukStatusActive
}

// IsMatured returns true if the sukuk has reached maturity
func (s *Sukuk) IsMatured() bool {
	return time.Now().After(s.MaturityDate) || s.Status == SukukStatusMatured
}

// CanAcceptInvestments returns true if the sukuk can accept new investments
func (s *Sukuk) CanAcceptInvestments() bool {
	return s.Status == SukukStatusActive && !s.IsMatured()
}

// GetActiveInvestments returns all active investments for this sukuk
func (s *Sukuk) GetActiveInvestments(db *gorm.DB) ([]Investment, error) {
	var investments []Investment
	err := db.Where("sukuk_id = ? AND status = ?", s.ID, InvestmentStatusActive).Find(&investments).Error
	return investments, err
}

// GetInvestorCount returns the number of unique active investors
func (s *Sukuk) GetInvestorCount(db *gorm.DB) (int64, error) {
	var count int64
	err := db.Model(&Investment{}).
		Where("sukuk_id = ? AND status = ?", s.ID, InvestmentStatusActive).
		Distinct("investor_address").
		Count(&count).Error
	return count, err
}

// GetPendingYields returns all pending yield claims for this sukuk
func (s *Sukuk) GetPendingYields(db *gorm.DB) ([]Yield, error) {
	var yields []Yield
	err := db.Where("sukuk_id = ? AND status = ?", s.ID, YieldStatusPending).Find(&yields).Error
	return yields, err
}

// GetPendingRedemptions returns all pending redemptions for this sukuk
func (s *Sukuk) GetPendingRedemptions(db *gorm.DB) ([]Redemption, error) {
	var redemptions []Redemption
	err := db.Where("sukuk_id = ? AND status = ?", s.ID, RedemptionStatusRequested).Find(&redemptions).Error
	return redemptions, err
}