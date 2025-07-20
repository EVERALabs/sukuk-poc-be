package models

import (
	"time"

	"gorm.io/gorm"
)

type Company struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	Name              string         `gorm:"size:255;not null" json:"name"`
	Description       string         `gorm:"type:text" json:"description"`
	Website           string         `gorm:"size:255" json:"website"`
	Industry          string         `gorm:"size:100" json:"industry"`
	Logo              string         `gorm:"size:255" json:"logo"`                  // File path to company logo
	LegalDocuments    string         `gorm:"type:text" json:"legal_documents"`      // JSON array of document paths
	WalletAddress     string         `gorm:"size:42;not null;uniqueIndex" json:"wallet_address"`
	Email             string         `gorm:"size:255;uniqueIndex" json:"email"`
	IsActive          bool           `gorm:"default:true" json:"is_active"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	SukukSeries []SukukSeries `gorm:"foreignKey:CompanyID" json:"sukuk_series,omitempty"`
}

func (Company) TableName() string {
	return "companies"
}