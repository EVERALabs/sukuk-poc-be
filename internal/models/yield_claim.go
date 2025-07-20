package models

import (
	"time"

	"gorm.io/gorm"
)

type YieldClaimStatus string

const (
	YieldClaimStatusPending   YieldClaimStatus = "pending"
	YieldClaimStatusClaimed   YieldClaimStatus = "claimed"
	YieldClaimStatusExpired   YieldClaimStatus = "expired"
)

type YieldClaim struct {
	ID               uint             `gorm:"primaryKey" json:"id"`
	SukukSeriesID    uint             `gorm:"not null;index" json:"sukuk_series_id"`
	InvestmentID     uint             `gorm:"not null;index" json:"investment_id"`
	InvestorAddress  string           `gorm:"size:42;not null;index" json:"investor_address"`
	YieldAmount      string           `gorm:"size:78;not null" json:"yield_amount"`             // Yield amount in IDRX
	PeriodStart      time.Time        `gorm:"not null" json:"period_start"`
	PeriodEnd        time.Time        `gorm:"not null" json:"period_end"`
	Status           YieldClaimStatus `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	TransactionHash  string           `gorm:"size:66;uniqueIndex" json:"transaction_hash"`      // Set when claimed
	BlockNumber      uint64           `json:"block_number"`                                     // Set when claimed
	ClaimedAt        *time.Time       `json:"claimed_at"`
	ExpiresAt        time.Time        `gorm:"not null" json:"expires_at"`                      // Yield claims can expire
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
	DeletedAt        gorm.DeletedAt   `gorm:"index" json:"-"`

	// Relationships
	SukukSeries SukukSeries `gorm:"foreignKey:SukukSeriesID" json:"sukuk_series,omitempty"`
	Investment  Investment  `gorm:"foreignKey:InvestmentID" json:"investment,omitempty"`
}

func (YieldClaim) TableName() string {
	return "yield_claims"
}