package models

import (
	"time"

	"gorm.io/gorm"
)

// SukukStatus represents the status of a Sukuk series
type SukukStatus string

const (
	SukukStatusActive   SukukStatus = "active"    // @Description Active and accepting investments
	SukukStatusMatured  SukukStatus = "matured"   // @Description Matured and no longer active
	SukukStatusSuspended SukukStatus = "suspended" // @Description Temporarily suspended
)

// SukukSeries represents a series of Sukuk tokens issued by a company
type SukukSeries struct {
	ID                   uint           `gorm:"primaryKey" json:"id" swaggertype:"integer" example:"1"`
	CompanyID            uint           `gorm:"not null;index" json:"company_id" swaggertype:"integer" example:"1"`
	Name                 string         `gorm:"size:255;not null" json:"name" swaggertype:"string" example:"Green Sukuk Series A"`
	Symbol               string         `gorm:"size:10;not null" json:"symbol" swaggertype:"string" example:"GSA"`
	Description          string         `gorm:"type:text" json:"description" swaggertype:"string" example:"Sustainable infrastructure financing sukuk"`
	TokenAddress         string         `gorm:"size:42;not null;uniqueIndex" json:"token_address" swaggertype:"string" example:"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"`
	TotalSupply          string         `gorm:"size:78;not null;default:'0'" json:"total_supply" swaggertype:"string" example:"1000000000000000000000000"`        // Use string for big numbers
	OutstandingSupply    string         `gorm:"size:78;not null;default:'0'" json:"outstanding_supply" swaggertype:"string" example:"500000000000000000000000"`  // Currently issued
	YieldRate            float64        `gorm:"type:decimal(5,4);not null" json:"yield_rate" swaggertype:"number" example:"0.085"`            // Annual yield rate (e.g., 0.0850 for 8.5%)
	MaturityDate         time.Time      `gorm:"not null" json:"maturity_date" swaggertype:"string" example:"2027-12-31T00:00:00Z"`
	PaymentFrequency     int            `gorm:"not null;default:4" json:"payment_frequency" swaggertype:"integer" example:"4"`              // Payments per year (quarterly = 4)
	MinInvestment        string         `gorm:"size:78;not null" json:"min_investment" swaggertype:"string" example:"1000000000000000000"`                   // Minimum investment amount
	MaxInvestment        string         `gorm:"size:78" json:"max_investment" swaggertype:"string" example:"100000000000000000000"`                            // Maximum investment amount (optional)
	Status               SukukStatus    `gorm:"type:varchar(20);not null;default:'active'" json:"status" swaggertype:"string" example:"active"`
	Prospectus           string         `gorm:"size:255" json:"prospectus" swaggertype:"string" example:"/uploads/prospectus/sukuk_1.pdf"`                               // PDF file path
	LegalDocuments       string         `gorm:"type:text" json:"legal_documents" swaggertype:"string" example:"[\"legal1.pdf\", \"legal2.pdf\"]"`                         // JSON array of document paths
	IsRedeemable         bool           `gorm:"default:true" json:"is_redeemable" swaggertype:"boolean" example:"true"`                        // Can investors redeem early
	CreatedAt            time.Time      `json:"created_at" swaggertype:"string" example:"2024-01-15T10:30:00Z"`
	UpdatedAt            time.Time      `json:"updated_at" swaggertype:"string" example:"2024-01-15T10:30:00Z"`
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