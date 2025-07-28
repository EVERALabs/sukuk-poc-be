package models

import (
	"time"
)

// RedemptionStatus represents the status of a redemption
type RedemptionStatus string

const (
	RedemptionStatusRequested RedemptionStatus = "requested"
	RedemptionStatusApproved  RedemptionStatus = "approved"
	RedemptionStatusRejected  RedemptionStatus = "rejected"
	RedemptionStatusCompleted RedemptionStatus = "completed"
)

// RedemptionRequest represents a comprehensive redemption with status
type RedemptionRequest struct {
	// Request Information
	RequestID       string    `json:"request_id"`
	User            string    `json:"user"`
	SukukAddress    string    `json:"sukuk_address"`
	Amount          string    `json:"amount"`
	PaymentToken    string    `json:"payment_token"`
	TotalSupply     string    `json:"total_supply"`
	RequestTxHash   string    `json:"request_tx_hash"`
	RequestTime     time.Time `json:"request_time"`
	RequestBlock    int64     `json:"request_block"`
	
	// Status and Approval Information
	Status              RedemptionStatus `json:"status"`
	ApprovalID          *string          `json:"approval_id,omitempty"`
	ApprovalTxHash      *string          `json:"approval_tx_hash,omitempty"`
	ApprovalTime        *time.Time       `json:"approval_time,omitempty"`
	ApprovalBlock       *int64           `json:"approval_block,omitempty"`
	ApprovedAmount      *string          `json:"approved_amount,omitempty"`
	
	// Metadata for UI/Business Logic
	Metadata            *SukukMetadata   `json:"metadata,omitempty"`
	CanApprove          bool             `json:"can_approve"`
	RequiresManagerAuth bool             `json:"requires_manager_auth"`
}

// RedemptionListResponse for API endpoints
type RedemptionListResponse struct {
	TotalCount   int                 `json:"total_count"`
	Redemptions  []RedemptionRequest `json:"redemptions"`
	StatusCounts map[string]int      `json:"status_counts"` // requested: 5, approved: 2, etc.
}

// RedemptionApprovalRequest for making approval calls
type RedemptionApprovalRequest struct {
	RequestID    string `json:"request_id" binding:"required"`
	SukukAddress string `json:"sukuk_address" binding:"required"`
	User         string `json:"user" binding:"required"`
	Amount       string `json:"amount" binding:"required"`
	PaymentToken string `json:"payment_token" binding:"required"`
	
	// Manager authorization
	ManagerAddress   string `json:"manager_address" binding:"required"`
	ManagerSignature string `json:"manager_signature,omitempty"`
}

// RedemptionStatsResponse for dashboard/overview
type RedemptionStatsResponse struct {
	TotalRequests        int    `json:"total_requests"`
	PendingRequests      int    `json:"pending_requests"`
	ApprovedRequests     int    `json:"approved_requests"`
	TotalRequestedAmount string `json:"total_requested_amount"`
	TotalApprovedAmount  string `json:"total_approved_amount"`
	
	// By Sukuk breakdown
	BySukuk map[string]RedemptionSukukStats `json:"by_sukuk"`
}

type RedemptionSukukStats struct {
	SukukAddress    string `json:"sukuk_address"`
	SukukCode       string `json:"sukuk_code"`
	RequestCount    int    `json:"request_count"`
	RequestedAmount string `json:"requested_amount"`
	ApprovedAmount  string `json:"approved_amount"`
}

// BlockchainCallRequest for making the actual approval transaction
type RedemptionBlockchainCall struct {
	SukukAddress string `json:"sukuk_address"`
	User         string `json:"user"`
	Amount       string `json:"amount"`
	PaymentToken string `json:"payment_token"`
	
	// Transaction details for the call
	GasLimit     string `json:"gas_limit,omitempty"`
	GasPrice     string `json:"gas_price,omitempty"`
	ManagerKey   string `json:"manager_key,omitempty"` // Private key or signer info
}