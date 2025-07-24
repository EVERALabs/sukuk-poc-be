package models

import (
	"time"
)

// SukukMetadataWithActivities extends SukukMetadata with latest blockchain activities
type SukukMetadataWithActivities struct {
	SukukMetadata
	LatestActivities []ActivityEvent `json:"latest_activities"`
}

// ActivityEvent represents a blockchain activity for a sukuk token
type ActivityEvent struct {
	Type         string    `json:"type"`          // "purchase" or "redemption_request"
	Address      string    `json:"address"`       // Buyer or User address
	Amount       string    `json:"amount"`        // Token amount
	TxHash       string    `json:"tx_hash"`       // Transaction hash
	Timestamp    time.Time `json:"timestamp"`     // Event timestamp
	SukukAddress string    `json:"sukuk_address"` // Sukuk contract address
}

// SukukMetadataListResponse represents the response for listing sukuk metadata with activities
type SukukMetadataListResponse struct {
	ID               uint            `json:"id"`
	ContractAddress  string          `json:"contract_address"`
	SukukCode        string          `json:"sukuk_code"`
	SukukTitle       string          `json:"sukuk_title"`
	SukukDeskripsi   string          `json:"sukuk_deskripsi"`
	Status           string          `json:"status"`
	LogoURL          string          `json:"logo_url"`
	Tenor            string          `json:"tenor"`
	ImbalHasil       string          `json:"imbal_hasil"`
	JatuhTempo       time.Time       `json:"jatuh_tempo"`
	KuotaNasional    float64         `json:"kuota_nasional"`
	MinimumPembelian float64         `json:"minimum_pembelian"`
	MetadataReady    bool            `json:"metadata_ready"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	LatestActivities []ActivityEvent `json:"latest_activities"`
}

// ToListResponse converts SukukMetadata to SukukMetadataListResponse
func (sm *SukukMetadata) ToListResponse() SukukMetadataListResponse {
	return SukukMetadataListResponse{
		ID:               sm.ID,
		ContractAddress:  sm.ContractAddress,
		SukukCode:        sm.SukukCode,
		SukukTitle:       sm.SukukTitle,
		SukukDeskripsi:   sm.SukukDeskripsi,
		Status:           sm.Status,
		LogoURL:          sm.LogoURL,
		Tenor:            sm.Tenor,
		ImbalHasil:       sm.ImbalHasil,
		JatuhTempo:       sm.JatuhTempo,
		KuotaNasional:    sm.KuotaNasional,
		MinimumPembelian: sm.MinimumPembelian,
		MetadataReady:    sm.MetadataReady,
		CreatedAt:        sm.CreatedAt,
		UpdatedAt:        sm.UpdatedAt,
		LatestActivities: []ActivityEvent{}, // Will be populated by service
	}
}