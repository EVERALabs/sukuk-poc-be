package models

import (
	"time"

	"gorm.io/gorm"
)

// SukukMetadata represents onchain sukuk data combined with offchain metadata
type SukukMetadata struct {
	// Primary Key
	ID uint `gorm:"primaryKey" json:"id"`

	// Onchain Data
	ContractAddress  string `gorm:"size:42;unique;not null" json:"contract_address"`
	TokenID          int64  `json:"token_id"`
	OwnerAddress     string `gorm:"size:42" json:"owner_address"`
	TransactionHash  string `gorm:"size:66" json:"transaction_hash"`
	BlockNumber      int64  `json:"block_number"`

	// Basic Info
	SukukCode      string `gorm:"size:20;not null" json:"sukuk_code"` // SR022-T5
	SukukTitle     string `gorm:"size:100" json:"sukuk_title"`               // Sukuk Ritel
	SukukDeskripsi string `gorm:"type:text" json:"sukuk_deskripsi"`          // Description
	Status         string `gorm:"size:20" json:"status"`                     // Berlangsung
	LogoURL        string `gorm:"size:255" json:"logo_url"`                  // Logo link

	// Main Features
	Tenor       string `gorm:"size:20" json:"tenor"`        // 5 Tahun
	ImbalHasil  string `gorm:"size:20" json:"imbal_hasil"`  // 6.55% / Tahun

	// Ketentuan SR022-T5
	PeriodePembelian     string    `gorm:"size:50" json:"periode_pembelian"`      // 16 Mei - 18 Jun 2025
	JatuhTempo           time.Time `json:"jatuh_tempo"`                            // 10 Jun 2030
	KuotaNasional        float64   `gorm:"type:decimal(30,2)" json:"kuota_nasional"` // Rp7,000,000,000,000
	PenerimaanKupon      string    `gorm:"size:20" json:"penerimaan_kupon"`       // Bulanan
	MinimumPembelian     float64   `gorm:"type:decimal(20,2)" json:"minimum_pembelian"` // Rp1,000,000
	TanggalBayarKupon    string    `gorm:"size:50" json:"tanggal_bayar_kupon"`    // 10 Setiap Bulan
	MaksimumPembelian    float64   `gorm:"type:decimal(20,2)" json:"maksimum_pembelian"` // Rp10,000,000,000
	KuponPertama         time.Time `json:"kupon_pertama"`                          // 11 Agustus 2025
	TipeKupon            string    `gorm:"size:20" json:"tipe_kupon"`             // Fixed Rate

	// Metadata Status
	MetadataReady bool `gorm:"default:false" json:"metadata_ready"`

	// Timestamps
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName overrides the table name
func (SukukMetadata) TableName() string {
	return "sukuk_metadata"
}


// SukukMetadataCreateRequest represents the request payload for creating sukuk metadata
type SukukMetadataCreateRequest struct {
	// Onchain Data
	ContractAddress  string `json:"contract_address" binding:"required"`
	TokenID          int64  `json:"token_id" binding:"required"`
	OwnerAddress     string `json:"owner_address" binding:"required"`
	TransactionHash  string `json:"transaction_hash"`
	BlockNumber      int64  `json:"block_number"`

	// Basic Info
	SukukCode      string `json:"sukuk_code" binding:"required"`
	SukukTitle     string `json:"sukuk_title"`
	SukukDeskripsi string `json:"sukuk_deskripsi"`
	Status         string `json:"status"`
	LogoURL        string `json:"logo_url"`

	// Main Features
	Tenor       string `json:"tenor"`
	ImbalHasil  string `json:"imbal_hasil"`

	// Ketentuan
	PeriodePembelian     string    `json:"periode_pembelian"`
	JatuhTempo           time.Time `json:"jatuh_tempo"`
	KuotaNasional        float64   `json:"kuota_nasional"`
	PenerimaanKupon      string    `json:"penerimaan_kupon"`
	MinimumPembelian     float64   `json:"minimum_pembelian"`
	TanggalBayarKupon    string    `json:"tanggal_bayar_kupon"`
	MaksimumPembelian    float64   `json:"maksimum_pembelian"`
	KuponPertama         time.Time `json:"kupon_pertama"`
	TipeKupon            string    `json:"tipe_kupon"`
}

// SukukMetadataUpdateRequest represents the request payload for updating sukuk metadata
// All fields are optional pointers to allow partial updates
type SukukMetadataUpdateRequest struct {
	// Basic Info
	SukukTitle     *string `json:"sukuk_title,omitempty"`
	SukukDeskripsi *string `json:"sukuk_deskripsi,omitempty"`
	Status         *string `json:"status,omitempty"`
	LogoURL        *string `json:"logo_url,omitempty"`

	// Main Features
	Tenor      *string `json:"tenor,omitempty"`
	ImbalHasil *string `json:"imbal_hasil,omitempty"`

	// Ketentuan
	PeriodePembelian     *string    `json:"periode_pembelian,omitempty"`
	JatuhTempo           *time.Time `json:"jatuh_tempo,omitempty"`
	KuotaNasional        *float64   `json:"kuota_nasional,omitempty"`
	PenerimaanKupon      *string    `json:"penerimaan_kupon,omitempty"`
	MinimumPembelian     *float64   `json:"minimum_pembelian,omitempty"`
	TanggalBayarKupon    *string    `json:"tanggal_bayar_kupon,omitempty"`
	MaksimumPembelian    *float64   `json:"maksimum_pembelian,omitempty"`
	KuponPertama         *time.Time `json:"kupon_pertama,omitempty"`
	TipeKupon            *string    `json:"tipe_kupon,omitempty"`
}

// SukukMetadataResponse represents the response for sukuk metadata
type SukukMetadataResponse struct {
	ID               uint `json:"id"`
	ContractAddress  string    `json:"contract_address"`
	TokenID          int64     `json:"token_id"`
	OwnerAddress     string    `json:"owner_address"`
	TransactionHash  string    `json:"transaction_hash"`
	BlockNumber      int64     `json:"block_number"`
	SukukCode        string    `json:"sukuk_code"`
	SukukTitle       string    `json:"sukuk_title"`
	SukukDeskripsi   string    `json:"sukuk_deskripsi"`
	Status           string    `json:"status"`
	LogoURL          string    `json:"logo_url"`
	Tenor            string    `json:"tenor"`
	ImbalHasil       string    `json:"imbal_hasil"`
	PeriodePembelian string    `json:"periode_pembelian"`
	JatuhTempo       time.Time `json:"jatuh_tempo"`
	KuotaNasional    float64   `json:"kuota_nasional"`
	PenerimaanKupon  string    `json:"penerimaan_kupon"`
	MinimumPembelian float64   `json:"minimum_pembelian"`
	TanggalBayarKupon string    `json:"tanggal_bayar_kupon"`
	MaksimumPembelian float64   `json:"maksimum_pembelian"`
	KuponPertama     time.Time `json:"kupon_pertama"`
	TipeKupon        string    `json:"tipe_kupon"`
	MetadataReady    bool      `json:"metadata_ready"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ToResponse converts SukukMetadata to SukukMetadataResponse
func (s *SukukMetadata) ToResponse() *SukukMetadataResponse {
	return &SukukMetadataResponse{
		ID:               s.ID,
		ContractAddress:  s.ContractAddress,
		TokenID:          s.TokenID,
		OwnerAddress:     s.OwnerAddress,
		TransactionHash:  s.TransactionHash,
		BlockNumber:      s.BlockNumber,
		SukukCode:        s.SukukCode,
		SukukTitle:       s.SukukTitle,
		SukukDeskripsi:   s.SukukDeskripsi,
		Status:           s.Status,
		LogoURL:          s.LogoURL,
		Tenor:            s.Tenor,
		ImbalHasil:       s.ImbalHasil,
		PeriodePembelian: s.PeriodePembelian,
		JatuhTempo:       s.JatuhTempo,
		KuotaNasional:    s.KuotaNasional,
		PenerimaanKupon:  s.PenerimaanKupon,
		MinimumPembelian: s.MinimumPembelian,
		TanggalBayarKupon: s.TanggalBayarKupon,
		MaksimumPembelian: s.MaksimumPembelian,
		KuponPertama:     s.KuponPertama,
		TipeKupon:        s.TipeKupon,
		MetadataReady:    s.MetadataReady,
		CreatedAt:        s.CreatedAt,
		UpdatedAt:        s.UpdatedAt,
	}
}