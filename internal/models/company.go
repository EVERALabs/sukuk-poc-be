package models

import (
	"time"

	"gorm.io/gorm"
)

// Company represents a partner company issuing Sukuk
type Company struct {
	ID                uint           `gorm:"primaryKey" json:"id" swaggertype:"integer" example:"1"`
	Name              string         `gorm:"size:255;not null" json:"name" swaggertype:"string" example:"PT Sukuk Indonesia"`
	Description       string         `gorm:"type:text" json:"description" swaggertype:"string" example:"Leading Indonesian sukuk issuer"`
	Website           string         `gorm:"size:255" json:"website" swaggertype:"string" example:"https://sukukindonesia.com"`
	Industry          string         `gorm:"size:100" json:"industry" swaggertype:"string" example:"Financial Services"`
	Logo              string         `gorm:"size:255" json:"logo" swaggertype:"string" example:"/uploads/logos/company_1.png"`                  // File path to company logo
	LegalDocuments    string         `gorm:"type:text" json:"legal_documents" swaggertype:"string" example:"[\"doc1.pdf\", \"doc2.pdf\"]"`      // JSON array of document paths
	WalletAddress     string         `gorm:"size:42;not null;uniqueIndex" json:"wallet_address" swaggertype:"string" example:"0x1234567890123456789012345678901234567890"`
	Email             string         `gorm:"size:255;uniqueIndex" json:"email" swaggertype:"string" example:"contact@sukukindonesia.com"`
	IsActive          bool           `gorm:"default:true" json:"is_active" swaggertype:"boolean" example:"true"`
	CreatedAt         time.Time      `json:"created_at" swaggertype:"string" example:"2024-01-15T10:30:00Z"`
	UpdatedAt         time.Time      `json:"updated_at" swaggertype:"string" example:"2024-01-15T10:30:00Z"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	SukukSeries []SukukSeries `gorm:"foreignKey:CompanyID" json:"sukuk_series,omitempty"`
}

func (Company) TableName() string {
	return "companies"
}