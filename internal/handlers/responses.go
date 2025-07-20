package handlers

import (
	"time"
	"github.com/kadzu/sukuk-poc-be/internal/models"
)

// Standard API response wrapper
type APIResponse struct {
	Message string      `json:"message,omitempty" example:"Operation completed successfully"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty" example:"Validation failed"`
}

// List response wrapper
type ListResponse struct {
	Data  interface{} `json:"data"`
	Count int         `json:"count" example:"5"`
	Meta  *MetaInfo   `json:"meta,omitempty"`
}

type MetaInfo struct {
	Total int `json:"total" example:"10"`
}

// Specific response structures for different endpoints

// Company responses
type CompanyResponse struct {
	Data models.Company `json:"data"`
}

type CompanyListResponse struct {
	Data  []models.Company `json:"data"`
	Count int              `json:"count" example:"5"`
}

type CompanySukuksResponse struct {
	Data  []models.SukukSeries `json:"data"`
	Count int                  `json:"count" example:"3"`
}

// Sukuk series responses
type SukukSeriesResponse struct {
	Data models.SukukSeries `json:"data"`
}

type SukukSeriesListResponse struct {
	Data  []models.SukukSeries `json:"data"`
	Count int                  `json:"count" example:"10"`
	Meta  MetaInfo             `json:"meta"`
}

type SukukMetricsResponse struct {
	Data SukukMetrics `json:"data"`
}

type SukukMetrics struct {
	TotalInvestors      int64  `json:"total_investors" example:"150"`
	TotalInvestment     string `json:"total_investment" example:"1000000000000000000000000"`
	PendingYields       int64  `json:"pending_yields" example:"25"`
	PendingRedemptions  int64  `json:"pending_redemptions" example:"5"`
}

type SukukHoldersResponse struct {
	Data  []models.Investment `json:"data"`
	Count int                 `json:"count" example:"150"`
}

// Investment responses
type InvestmentResponse struct {
	Data models.Investment `json:"data"`
}

type InvestmentListResponse struct {
	Data  []models.Investment `json:"data"`
	Count int                 `json:"count" example:"100"`
}

// Portfolio responses
type PortfolioResponse struct {
	Data PortfolioData `json:"data"`
}

type PortfolioData struct {
	Address        string                `json:"address" example:"0x1234567890123456789012345678901234567890"`
	Summary        PortfolioSummary      `json:"summary"`
	Investments    []models.Investment   `json:"investments"`
	PendingYields  []models.YieldClaim   `json:"pending_yields"`
	Redemptions    []models.Redemption   `json:"redemptions"`
}

type PortfolioSummary struct {
	TotalInvestments   int `json:"total_investments" example:"5"`
	PendingYields      int `json:"pending_yields" example:"3"`
	TotalRedemptions   int `json:"total_redemptions" example:"2"`
}

// Yield claim responses
type YieldClaimResponse struct {
	Data models.YieldClaim `json:"data"`
}

type YieldClaimListResponse struct {
	Data  []models.YieldClaim `json:"data"`
	Count int                 `json:"count" example:"50"`
}

// Redemption responses
type RedemptionResponse struct {
	Data models.Redemption `json:"data"`
}

type RedemptionListResponse struct {
	Data  []models.Redemption `json:"data"`
	Count int                 `json:"count" example:"30"`
}

// Analytics responses
type PlatformStatsResponse struct {
	Data PlatformStats `json:"data"`
}

type PlatformStats struct {
	TotalCompanies     int64 `json:"total_companies" example:"10"`
	TotalSukukSeries   int64 `json:"total_sukuk_series" example:"25"`
	TotalInvestments   int64 `json:"total_investments" example:"500"`
	UniqueInvestors    int64 `json:"unique_investors" example:"250"`
	PendingYields      int64 `json:"pending_yields" example:"50"`
	PendingRedemptions int64 `json:"pending_redemptions" example:"15"`
}

type VaultBalanceResponse struct {
	Data VaultBalance `json:"data"`
}

type VaultBalance struct {
	SeriesID        uint   `json:"series_id" example:"1"`
	SeriesName      string `json:"series_name" example:"Green Sukuk Series A"`
	VaultBalance    string `json:"vault_balance" example:"500000000000000000000000"`
	TotalSupply     string `json:"total_supply" example:"1000000000000000000000000"`
	UtilizationRate string `json:"utilization_rate" example:"0.5"`
}

// Event responses
type EventListResponse struct {
	Data  []models.Event `json:"data"`
	Count int            `json:"count" example:"10"`
}

type EventWebhookResponse struct {
	Message string `json:"message" example:"Event processed successfully"`
	EventID uint   `json:"event_id" example:"123"`
}

// File upload responses
type FileUploadResponse struct {
	Message  string `json:"message" example:"File uploaded successfully"`
	Filename string `json:"filename" example:"company_logo_1.png"`
	URL      string `json:"url" example:"/uploads/logos/company_logo_1.png"`
}

// Error response
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid request data"`
}

// Health check responses
type HealthResponse struct {
	Status      string          `json:"status" example:"healthy"`
	Timestamp   time.Time       `json:"timestamp" example:"2024-01-15T10:30:00Z"`
	Version     string          `json:"version" example:"1.0.0"`
	Environment string          `json:"environment" example:"development"`
	Checks      []CheckResult   `json:"checks"`
}

