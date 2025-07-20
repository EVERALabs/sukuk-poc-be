package models

import (
	"time"

	"gorm.io/gorm"
)

type RedemptionStatus string

const (
	RedemptionStatusRequested RedemptionStatus = "requested"
	RedemptionStatusApproved  RedemptionStatus = "approved"
	RedemptionStatusRejected  RedemptionStatus = "rejected"
	RedemptionStatusCompleted RedemptionStatus = "completed"
	RedemptionStatusCancelled RedemptionStatus = "cancelled"
)

type Redemption struct {
	ID               uint             `gorm:"primaryKey" json:"id"`
	SukukSeriesID    uint             `gorm:"not null;index" json:"sukuk_series_id"`
	InvestmentID     uint             `gorm:"not null;index" json:"investment_id"`
	InvestorAddress  string           `gorm:"size:42;not null;index" json:"investor_address"`
	TokenAmount      string           `gorm:"size:78;not null" json:"token_amount"`             // Amount of tokens to redeem
	RedemptionAmount string           `gorm:"size:78;not null" json:"redemption_amount"`        // IDRX amount to be returned
	Status           RedemptionStatus `gorm:"type:varchar(20);not null;default:'requested'" json:"status"`
	RequestReason    string           `gorm:"type:text" json:"request_reason"`                  // Why investor wants to redeem
	ApprovalNotes    string           `gorm:"type:text" json:"approval_notes"`                  // Company's approval/rejection notes
	TransactionHash  string           `gorm:"size:66;uniqueIndex" json:"transaction_hash"`      // Set when completed
	BlockNumber      uint64           `json:"block_number"`                                     // Set when completed
	RequestedAt      time.Time        `gorm:"not null" json:"requested_at"`
	ApprovedAt       *time.Time       `json:"approved_at"`
	RejectedAt       *time.Time       `json:"rejected_at"`
	RejectionReason  string           `gorm:"type:text" json:"rejection_reason"`
	CompletedAt      *time.Time       `json:"completed_at")`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedAt        time.Time        `json:"updated_at"`
	DeletedAt        gorm.DeletedAt   `gorm:"index" json:"-"`

	// Relationships
	SukukSeries SukukSeries `gorm:"foreignKey:SukukSeriesID" json:"sukuk_series,omitempty"`
	Investment  Investment  `gorm:"foreignKey:InvestmentID" json:"investment,omitempty"`
}

func (Redemption) TableName() string {
	return "redemptions"
}