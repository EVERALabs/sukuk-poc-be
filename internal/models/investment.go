package models

import (
	"time"

	"gorm.io/gorm"
)

type InvestmentStatus string

const (
	InvestmentStatusActive    InvestmentStatus = "active"
	InvestmentStatusRedeemed  InvestmentStatus = "redeemed"
	InvestmentStatusMatured   InvestmentStatus = "matured"
)

type Investment struct {
	ID               uint             `gorm:"primaryKey" json:"id"`
	SukukSeriesID    uint             `gorm:"not null;index" json:"sukuk_series_id"`
	InvestorAddress  string           `gorm:"size:42;not null;index" json:"investor_address"`
	InvestorEmail    string           `gorm:"size:255;index" json:"investor_email"`
	Amount           string           `gorm:"size:78;not null" json:"amount"`                  // Investment amount in IDRX
	TokenAmount      string           `gorm:"size:78;not null" json:"token_amount"`            // Equivalent Sukuk tokens received
	Status           InvestmentStatus `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	TransactionHash  string           `gorm:"size:66;not null;uniqueIndex" json:"transaction_hash"`
	BlockNumber      uint64           `gorm:"not null" json:"block_number"`
	InvestmentDate   time.Time        `gorm:"not null" json:"investment_date"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
	DeletedAt        gorm.DeletedAt   `gorm:"index" json:"-"`

	// Relationships
	SukukSeries SukukSeries  `gorm:"foreignKey:SukukSeriesID" json:"sukuk_series,omitempty"`
	YieldClaims []YieldClaim `gorm:"foreignKey:InvestmentID" json:"yield_claims,omitempty"`
	Redemptions []Redemption `gorm:"foreignKey:InvestmentID" json:"redemptions,omitempty"`
}

func (Investment) TableName() string {
	return "investments"
}