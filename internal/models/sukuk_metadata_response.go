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
	SukukCode    string    `json:"sukuk_code"`    // Sukuk symbol/code (e.g., "SITI")
	SukukTitle   string    `json:"sukuk_title"`   // Sukuk name/title
}

// SukukMetadataListResponse represents the response for listing sukuk metadata with activities
type SukukMetadataListResponse struct {
	ID               uint            `json:"id"`
	ContractAddress  string          `json:"contract_address"`
	TokenID          int64           `json:"token_id"`
	OwnerAddress     string          `json:"owner_address"`
	TransactionHash  string          `json:"transaction_hash"`
	BlockNumber      int64           `json:"block_number"`
	SukukCode        string          `json:"sukuk_code"`
	SukukTitle       string          `json:"sukuk_title"`
	SukukDeskripsi   string          `json:"sukuk_deskripsi"`
	Status           string          `json:"status"`
	LogoURL          string          `json:"logo_url"`
	Tenor            string          `json:"tenor"`
	ImbalHasil       string          `json:"imbal_hasil"`
	PeriodePembelian string          `json:"periode_pembelian"`
	JatuhTempo       time.Time       `json:"jatuh_tempo"`
	KuotaNasional    float64         `json:"kuota_nasional"`
	PenerimaanKupon  string          `json:"penerimaan_kupon"`
	MinimumPembelian float64         `json:"minimum_pembelian"`
	TanggalBayarKupon string         `json:"tanggal_bayar_kupon"`
	MaksimumPembelian float64        `json:"maksimum_pembelian"`
	KuponPertama     time.Time       `json:"kupon_pertama"`
	TipeKupon        string          `json:"tipe_kupon"`
	MetadataReady    bool            `json:"metadata_ready"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	LatestActivities []ActivityEvent `json:"latest_activities"`
}

// ToListResponse converts SukukMetadata to SukukMetadataListResponse
func (sm *SukukMetadata) ToListResponse() SukukMetadataListResponse {
	return SukukMetadataListResponse{
		ID:                sm.ID,
		ContractAddress:   sm.ContractAddress,
		TokenID:           sm.TokenID,
		OwnerAddress:      sm.OwnerAddress,
		TransactionHash:   sm.TransactionHash,
		BlockNumber:       sm.BlockNumber,
		SukukCode:         sm.SukukCode,
		SukukTitle:        sm.SukukTitle,
		SukukDeskripsi:    sm.SukukDeskripsi,
		Status:            sm.Status,
		LogoURL:           sm.LogoURL,
		Tenor:             sm.Tenor,
		ImbalHasil:        sm.ImbalHasil,
		PeriodePembelian:  sm.PeriodePembelian,
		JatuhTempo:        sm.JatuhTempo,
		KuotaNasional:     sm.KuotaNasional,
		PenerimaanKupon:   sm.PenerimaanKupon,
		MinimumPembelian:  sm.MinimumPembelian,
		TanggalBayarKupon: sm.TanggalBayarKupon,
		MaksimumPembelian: sm.MaksimumPembelian,
		KuponPertama:      sm.KuponPertama,
		TipeKupon:         sm.TipeKupon,
		MetadataReady:     sm.MetadataReady,
		CreatedAt:         sm.CreatedAt,
		UpdatedAt:         sm.UpdatedAt,
		LatestActivities:  []ActivityEvent{}, // Will be populated by service
	}
}