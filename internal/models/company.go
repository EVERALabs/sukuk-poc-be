package models

import (
	"time"

	"gorm.io/gorm"
)

// Company represents a sukuk issuing company
type Company struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"size:255;not null" json:"name"`
	Description   string    `gorm:"type:text" json:"description"`
	Website       string    `gorm:"size:255" json:"website"`
	Industry      string    `gorm:"size:100" json:"industry"`
	Email         string    `gorm:"size:255;uniqueIndex;not null" json:"email"`
	WalletAddress string    `gorm:"size:42;uniqueIndex;not null" json:"wallet_address"`
	Logo          string    `gorm:"size:500" json:"logo"`
	IsActive      bool      `gorm:"default:true;not null" json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Sukuks []Sukuk `gorm:"foreignKey:CompanyID" json:"sukuks,omitempty"`
}

// TableName returns the table name for Company model
func (Company) TableName() string {
	return "companies"
}

// BeforeCreate hook to normalize wallet address
func (c *Company) BeforeCreate(tx *gorm.DB) error {
	c.WalletAddress = normalizeAddress(c.WalletAddress)
	return nil
}

// BeforeUpdate hook to normalize wallet address
func (c *Company) BeforeUpdate(tx *gorm.DB) error {
	if c.WalletAddress != "" {
		c.WalletAddress = normalizeAddress(c.WalletAddress)
	}
	return nil
}

// GetActiveSukuks returns only active sukuk series for this company
func (c *Company) GetActiveSukuks(db *gorm.DB) ([]Sukuk, error) {
	var sukuks []Sukuk
	err := db.Where("company_id = ? AND status = ?", c.ID, SukukStatusActive).Find(&sukuks).Error
	return sukuks, err
}

// GetSukukCount returns the total number of sukuk series for this company
func (c *Company) GetSukukCount(db *gorm.DB) (int64, error) {
	var count int64
	err := db.Model(&Sukuk{}).Where("company_id = ?", c.ID).Count(&count).Error
	return count, err
}