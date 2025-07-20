package models

import (
	"time"

	"gorm.io/gorm"
)

type SukukStatus string

const (
	SukukStatusActive   SukukStatus = "active"
	SukukStatusMatured  SukukStatus = "matured"
	SukukStatusSuspended SukukStatus = "suspended"
)

type SukukSeries struct {
	ID                   uint           `gorm:"primaryKey" json:"id"`
	CompanyID            uint           `gorm:"not null;index" json:"company_id"`
	Name                 string         `gorm:"size:255;not null" json:"name"`
	Symbol               string         `gorm:"size:10;not null" json:"symbol"`
	Description          string         `gorm:"type:text" json:"description"`
	TokenAddress         string         `gorm:"size:42;not null;uniqueIndex" json:"token_address"`
	TotalSupply          string         `gorm:"size:78;not null;default:'0'" json:"total_supply"`        // Use string for big numbers
	OutstandingSupply    string         `gorm:"size:78;not null;default:'0'" json:"outstanding_supply"`  // Currently issued
	YieldRate            float64        `gorm:"type:decimal(5,4);not null" json:"yield_rate"`            // Annual yield rate (e.g., 0.0850 for 8.5%)
	MaturityDate         time.Time      `gorm:"not null" json:"maturity_date"`
	PaymentFrequency     int            `gorm:"not null;default:4" json:"payment_frequency"`              // Payments per year (quarterly = 4)
	MinInvestment        string         `gorm:"size:78;not null" json:"min_investment"`                   // Minimum investment amount
	MaxInvestment        string         `gorm:"size:78" json:"max_investment"`                            // Maximum investment amount (optional)
	Status               SukukStatus    `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	Prospectus           string         `gorm:"size:255" json:"prospectus"`                               // PDF file path
	LegalDocuments       string         `gorm:"type:text" json:"legal_documents"`                         // JSON array of document paths
	IsRedeemable         bool           `gorm:"default:true" json:"is_redeemable"`                        // Can investors redeem early
	CreatedAt            time.Time      `json:"created_at"`
	UpdatedAt            time.Time      `json:"updated_at"`
	DeletedAt            gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Company     Company      `gorm:"foreignKey:CompanyID" json:"company,omitempty"`
	Investments []Investment `gorm:"foreignKey:SukukSeriesID" json:"investments,omitempty"`
	YieldClaims []YieldClaim `gorm:"foreignKey:SukukSeriesID" json:"yield_claims,omitempty"`
	Redemptions []Redemption `gorm:"foreignKey:SukukSeriesID" json:"redemptions,omitempty"`
}

func (SukukSeries) TableName() string {
	return "sukuk_series"
}