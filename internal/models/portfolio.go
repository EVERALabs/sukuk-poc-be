package models

import (
	"time"
)

// Portfolio response structures for API endpoints

// PortfolioResponse represents a user's complete portfolio
type PortfolioResponse struct {
	Address      string            `json:"address"`
	TotalHoldings int              `json:"total_holdings"`
	Holdings     []SukukHolding    `json:"holdings"`
	TotalValue   string            `json:"total_value,omitempty"`    // Total portfolio value in USD/base currency
	Summary      PortfolioSummary  `json:"summary"`
}

// SukukHolding represents user's holding in a specific sukuk
type SukukHolding struct {
	SukukAddress           string               `json:"sukuk_address"`
	Balance                string               `json:"balance"`                    // Current token balance
	ClaimableYield         string               `json:"claimable_yield"`           // Available yield to claim
	TotalYieldClaimed      string               `json:"total_yield_claimed"`       // Total yield claimed historically
	UnclaimedDistributions []int64              `json:"unclaimed_distribution_ids"` // Distribution IDs available for claiming
	LastActivity           *time.Time           `json:"last_activity,omitempty"`   // Last purchase/redemption
	Metadata               *SukukMetadata       `json:"metadata,omitempty"`        // Sukuk details
	YieldHistory           []YieldDistribution  `json:"yield_history,omitempty"`   // Recent yield distributions
}

// PortfolioSummary provides aggregate portfolio statistics
type PortfolioSummary struct {
	TotalSukukCount      int    `json:"total_sukuk_count"`
	TotalClaimableYield  string `json:"total_claimable_yield"`
	TotalYieldClaimed    string `json:"total_yield_claimed"`
	ActiveSukukCount     int    `json:"active_sukuk_count"`     // Sukuk with non-zero balance
	MaturedSukukCount    int    `json:"matured_sukuk_count"`    // Sukuk that have matured
}

// YieldClaimsResponse represents available yield claims for a user
type YieldClaimsResponse struct {
	Address      string              `json:"address"`
	TotalClaims  int                 `json:"total_claims"`
	Claims       []YieldClaimDetail  `json:"claims"`
	TotalAmount  string              `json:"total_amount"`   // Total claimable across all sukuk
}

// YieldClaimDetail represents claimable yield for a specific sukuk
type YieldClaimDetail struct {
	SukukAddress         string    `json:"sukuk_address"`
	ClaimableAmount      string    `json:"claimable_amount"`
	LastDistribution     *time.Time `json:"last_distribution,omitempty"`
	DistributionCount    int       `json:"distribution_count"`
	UserBalance          string    `json:"user_balance"`
	UnclaimedDistributions []int64 `json:"unclaimed_distribution_ids"` // Distribution IDs available for claiming
	Metadata             *SukukMetadata `json:"metadata,omitempty"`
}

// YieldDistribution represents a yield distribution event
type YieldDistribution struct {
	ID             string    `json:"id"`
	SukukAddress   string    `json:"sukuk_address"`
	DistributionId int64     `json:"distribution_id"`  // Required for claiming yields
	Amount         string    `json:"amount"`
	Timestamp      time.Time `json:"timestamp"`
	TxHash         string    `json:"tx_hash"`
	BlockNumber    int64     `json:"block_number"`
}

// YieldClaim represents a yield claim event
type YieldClaim struct {
	ID           string    `json:"id"`
	User         string    `json:"user"`
	SukukAddress string    `json:"sukuk_address"`
	Amount       string    `json:"amount"`
	Timestamp    time.Time `json:"timestamp"`
	TxHash       string    `json:"tx_hash"`
	BlockNumber  int64     `json:"block_number"`
}

// TransactionHistory represents combined transaction history for a user
type TransactionHistoryResponse struct {
	Address      string              `json:"address"`
	TotalCount   int                 `json:"total_count"`
	Transactions []TransactionEvent  `json:"transactions"`
}

// TransactionEvent represents any blockchain event related to the user
type TransactionEvent struct {
	Type         string    `json:"type"`           // "purchase", "redemption", "yield_claim", "yield_distribution"
	SukukAddress string    `json:"sukuk_address"`
	Amount       string    `json:"amount"`
	TxHash       string    `json:"tx_hash"`
	Timestamp    time.Time `json:"timestamp"`
	BlockNumber  int64     `json:"block_number"`
	Status       string    `json:"status,omitempty"`    // "pending", "confirmed", "failed"
	Details      map[string]interface{} `json:"details,omitempty"` // Additional event-specific data
}

// IndexerTableInfo represents discovered indexer table information
type IndexerTableInfo struct {
	EventType    string `json:"event_type"`
	TableName    string `json:"table_name"`
	HashPrefix   string `json:"hash_prefix"`
	RowCount     int64  `json:"row_count,omitempty"`
	LastUpdated  *time.Time `json:"last_updated,omitempty"`
}

// IndexerTablesResponse represents the debug response for indexer tables
type IndexerTablesResponse struct {
	TotalTables      int                 `json:"total_tables"`
	AvailableEvents  []string            `json:"available_events"`
	Tables           []IndexerTableInfo  `json:"tables"`
	LatestTables     map[string]string   `json:"latest_tables"`    // event_type -> table_name mapping
}

// HoldingCalculation represents intermediate calculation data
type HoldingCalculation struct {
	UserAddress      string  `json:"user_address"`
	SukukAddress     string  `json:"sukuk_address"`
	CurrentBalance   string  `json:"current_balance"`
	PurchaseHistory  []PurchaseEvent  `json:"purchase_history,omitempty"`
	RedemptionHistory []RedemptionEvent `json:"redemption_history,omitempty"`
	DistributionShare float64 `json:"distribution_share"`  // User's share of distributions (0.0-1.0)
}

// PurchaseEvent represents a sukuk purchase
type PurchaseEvent struct {
	ID           string    `json:"id"`
	Buyer        string    `json:"buyer"`
	SukukAddress string    `json:"sukuk_address"`
	Amount       string    `json:"amount"`
	PaymentToken string    `json:"payment_token"`
	Timestamp    time.Time `json:"timestamp"`
	TxHash       string    `json:"tx_hash"`
	BlockNumber  int64     `json:"block_number"`
}

// RedemptionEvent represents a sukuk redemption request
type RedemptionEvent struct {
	ID           string    `json:"id"`
	User         string    `json:"user"`
	SukukAddress string    `json:"sukuk_address"`
	Amount       string    `json:"amount"`
	PaymentToken string    `json:"payment_token"`
	Timestamp    time.Time `json:"timestamp"`
	TxHash       string    `json:"tx_hash"`
	BlockNumber  int64     `json:"block_number"`
	TotalSupply  string    `json:"total_supply,omitempty"`
}

// SnapshotEvent represents a balance snapshot for yield distribution
type SnapshotEvent struct {
	ID           string    `json:"id"`
	SukukAddress string    `json:"sukuk_address"`
	SnapshotId   string    `json:"snapshot_id"`
	TotalSupply  string    `json:"total_supply"`
	HolderCount  int64     `json:"holder_count"`
	EligibleCount int64    `json:"eligible_count"`
	Timestamp    time.Time `json:"timestamp"`
	TxHash       string    `json:"tx_hash"`
	BlockNumber  int64     `json:"block_number"`
}

// HolderBalance represents a user's balance at a specific time
type HolderBalance struct {
	ID           string    `json:"id"`
	SukukAddress string    `json:"sukuk_address"`
	User         string    `json:"user"`
	Balance      string    `json:"balance"`
	Timestamp    time.Time `json:"timestamp"`
	TxHash       string    `json:"tx_hash"`
	BlockNumber  int64     `json:"block_number"`
}