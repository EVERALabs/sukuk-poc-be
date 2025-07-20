package models

import (
	"time"

	"gorm.io/gorm"
)

// InvestmentStatus represents the status of an investment
type InvestmentStatus string

const (
	InvestmentStatusActive    InvestmentStatus = "active"    // @Description Investment is active and earning yield
	InvestmentStatusRedeemed  InvestmentStatus = "redeemed"   // @Description Investment has been redeemed
	InvestmentStatusMatured   InvestmentStatus = "matured"    // @Description Investment has reached maturity
)

// Investment represents an investment made in a Sukuk series
type Investment struct {
	ID               uint             `gorm:"primaryKey" json:"id" swaggertype:"integer" example:"1"`
	SukukSeriesID    uint             `gorm:"not null;index" json:"sukuk_series_id" swaggertype:"integer" example:"1"`
	InvestorAddress  string           `gorm:"size:42;not null;index" json:"investor_address" swaggertype:"string" example:"0x1234567890123456789012345678901234567890"`
	InvestorEmail    string           `gorm:"size:255;index" json:"investor_email" swaggertype:"string" example:"investor@example.com"`
	Amount           string           `gorm:"size:78;not null" json:"amount" swaggertype:"string" example:"1000000000000000000000"`                  // Investment amount in IDRX
	TokenAmount      string           `gorm:"size:78;not null" json:"token_amount" swaggertype:"string" example:"1000000000000000000000"`            // Equivalent Sukuk tokens received
	Status           InvestmentStatus `gorm:"type:varchar(20);not null;default:'active'" json:"status" swaggertype:"string" example:"active"`
	TransactionHash  string           `gorm:"size:66;not null;uniqueIndex" json:"transaction_hash" swaggertype:"string" example:"0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"`
	BlockNumber      uint64           `gorm:"not null" json:"block_number" swaggertype:"integer" example:"12345678"`
	InvestmentDate   time.Time        `gorm:"not null" json:"investment_date" swaggertype:"string" example:"2024-01-15T10:30:00Z"`
	CreatedAt        time.Time        `json:"created_at" swaggertype:"string" example:"2024-01-15T10:30:00Z"`
	UpdatedAt        time.Time        `json:"updated_at" swaggertype:"string" example:"2024-01-15T10:30:00Z"`
	DeletedAt        gorm.DeletedAt   `gorm:"index" json:"-"`

	// Relationships
	SukukSeries SukukSeries  `gorm:"foreignKey:SukukSeriesID" json:"sukuk_series,omitempty"`
	YieldClaims []YieldClaim `gorm:"foreignKey:InvestmentID" json:"yield_claims,omitempty"`
	Redemptions []Redemption `gorm:"foreignKey:InvestmentID" json:"redemptions,omitempty"`
}

func (Investment) TableName() string {
	return "investments"
}