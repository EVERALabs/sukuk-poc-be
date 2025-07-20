package database

import (
	"log"
	"time"

	"sukuk-be/internal/models"

	"gorm.io/gorm"
)

// SeedData populates the database with sample data for testing
func SeedData(db *gorm.DB) error {
	log.Println("Starting database seeding...")

	// Check if data already exists
	var count int64
	db.Model(&models.Company{}).Count(&count)
	if count > 0 {
		log.Println("Database already contains data, skipping seed")
		return nil
	}

	// Create sample companies
	companies := []models.Company{
		{
			Name:          "PT PLN (Persero)",
			Description:   "Indonesian state-owned electricity company providing power generation and distribution across Indonesia.",
			Website:       "https://www.pln.co.id",
			Industry:      "Energy",
			Email:         "investor@pln.co.id",
			WalletAddress: "0x1234567890123456789012345678901234567890",
			IsActive:      true,
		},
		{
			Name:          "PT Antam Tbk",
			Description:   "Indonesian state-owned mining company focused on gold, silver, and other precious metals.",
			Website:       "https://www.antam.com",
			Industry:      "Mining",
			Email:         "ir@antam.com",
			WalletAddress: "0x2345678901234567890123456789012345678901",
			IsActive:      true,
		},
		{
			Name:          "PT Telkom Indonesia",
			Description:   "Indonesia's largest telecommunications company providing internet, mobile, and digital services.",
			Website:       "https://www.telkom.co.id",
			Industry:      "Telecommunications",
			Email:         "investor@telkom.co.id",
			WalletAddress: "0x3456789012345678901234567890123456789012",
			IsActive:      true,
		},
	}

	for i := range companies {
		if err := db.Create(&companies[i]).Error; err != nil {
			return err
		}
	}
	log.Printf("Created %d companies", len(companies))

	// Create sample Sukuk series
	sukukSeries := []models.SukukSeries{
		{
			CompanyID:         companies[0].ID, // PLN
			Name:              "PLN Green Energy Sukuk 2024-A",
			Symbol:            "PLN24A",
			Description:       "Islamic bonds for financing renewable energy projects across Indonesia",
			TokenAddress:      "0x4567890123456789012345678901234567890123",
			TotalSupply:       "1000000000000000000000000000", // 1B tokens
			OutstandingSupply: "500000000000000000000000000",  // 500M tokens
			YieldRate:         8.5,
			MaturityDate:      time.Date(2027, 12, 31, 0, 0, 0, 0, time.UTC),
			PaymentFrequency:  4,                             // Quarterly
			MinInvestment:     "1000000000000000000000000",   // 1M IDRX
			MaxInvestment:     "100000000000000000000000000", // 100M IDRX
			Status:            "active",
			Prospectus:        "/uploads/prospectus/pln_green_energy_2024.pdf",
			IsRedeemable:      true,
		},
		{
			CompanyID:         companies[1].ID, // Antam
			Name:              "Antam Gold Mining Sukuk 2024-B",
			Symbol:            "ANTM24B",
			Description:       "Sharia-compliant financing for sustainable gold mining operations",
			TokenAddress:      "0x5678901234567890123456789012345678901234",
			TotalSupply:       "500000000000000000000000000", // 500M tokens
			OutstandingSupply: "250000000000000000000000000", // 250M tokens
			YieldRate:         9.0,
			MaturityDate:      time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC),
			PaymentFrequency:  4,                            // Quarterly
			MinInvestment:     "5000000000000000000000000",  // 5M IDRX
			MaxInvestment:     "50000000000000000000000000", // 50M IDRX
			Status:            "active",
			Prospectus:        "/uploads/prospectus/antam_gold_mining_2024.pdf",
			IsRedeemable:      true,
		},
		{
			CompanyID:         companies[2].ID, // Telkom
			Name:              "Telkom Digital Infrastructure Sukuk 2024-C",
			Symbol:            "TLKM24C",
			Description:       "Islamic financing for 5G network expansion and digital infrastructure development",
			TokenAddress:      "",                            // Not deployed yet
			TotalSupply:       "750000000000000000000000000", // 750M tokens
			OutstandingSupply: "0",                           // Not issued yet
			YieldRate:         7.5,
			MaturityDate:      time.Date(2028, 3, 31, 0, 0, 0, 0, time.UTC),
			PaymentFrequency:  4,                            // Quarterly
			MinInvestment:     "2000000000000000000000000",  // 2M IDRX
			MaxInvestment:     "75000000000000000000000000", // 75M IDRX
			Status:            "planned",                    // Not active yet
			Prospectus:        "",
			IsRedeemable:      true,
		},
	}

	for i := range sukukSeries {
		if err := db.Create(&sukukSeries[i]).Error; err != nil {
			return err
		}
	}
	log.Printf("Created %d sukuk series", len(sukukSeries))

	// Create sample investments (for PLN and Antam series only)
	investments := []models.Investment{
		{
			SukukSeriesID:   sukukSeries[0].ID, // PLN
			InvestorAddress: "0xabc1234567890123456789012345678901234567",
			InvestorEmail:   "investor1@example.com",
			Amount:          "10000000000000000000000000", // 10M IDRX
			TokenAmount:     "10000000000000000000000000", // 10M PLN24A tokens
			Status:          "active",
			TransactionHash: "0xdef1234567890123456789012345678901234567890123456789012345678901",
			BlockNumber:     123456,
			InvestmentDate:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
		{
			SukukSeriesID:   sukukSeries[0].ID, // PLN
			InvestorAddress: "0xbcd2345678901234567890123456789012345678",
			InvestorEmail:   "investor2@example.com",
			Amount:          "25000000000000000000000000", // 25M IDRX
			TokenAmount:     "25000000000000000000000000", // 25M PLN24A tokens
			Status:          "active",
			TransactionHash: "0xeef2345678901234567890123456789012345678901234567890123456789012",
			BlockNumber:     123789,
			InvestmentDate:  time.Date(2024, 2, 10, 14, 20, 0, 0, time.UTC),
		},
		{
			SukukSeriesID:   sukukSeries[1].ID, // Antam
			InvestorAddress: "0xabc1234567890123456789012345678901234567",
			InvestorEmail:   "investor1@example.com",
			Amount:          "15000000000000000000000000", // 15M IDRX
			TokenAmount:     "15000000000000000000000000", // 15M ANTM24B tokens
			Status:          "active",
			TransactionHash: "0xfff3456789012345678901234567890123456789012345678901234567890123",
			BlockNumber:     124567,
			InvestmentDate:  time.Date(2024, 3, 5, 9, 15, 0, 0, time.UTC),
		},
	}

	for i := range investments {
		if err := db.Create(&investments[i]).Error; err != nil {
			return err
		}
	}
	log.Printf("Created %d investments", len(investments))

	// Create sample yield claims
	yieldClaims := []models.YieldClaim{
		{
			SukukSeriesID:   sukukSeries[0].ID, // PLN
			InvestmentID:    investments[0].ID,
			InvestorAddress: "0xabc1234567890123456789012345678901234567",
			YieldAmount:     "212500000000000000000000", // ~2.125% quarterly yield
			PeriodStart:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			PeriodEnd:       time.Date(2024, 3, 31, 23, 59, 59, 0, time.UTC),
			Status:          "pending",
			ExpiresAt:       time.Date(2024, 4, 30, 23, 59, 59, 0, time.UTC),
		},
		{
			SukukSeriesID:   sukukSeries[1].ID, // Antam
			InvestmentID:    investments[2].ID,
			InvestorAddress: "0xabc1234567890123456789012345678901234567",
			YieldAmount:     "337500000000000000000000", // ~2.25% quarterly yield
			PeriodStart:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			PeriodEnd:       time.Date(2024, 3, 31, 23, 59, 59, 0, time.UTC),
			Status:          "claimed",
			TransactionHash: "0x1112345678901234567890123456789012345678901234567890123456789012",
			BlockNumber:     125678,
			ClaimedAt:       func() *time.Time { t := time.Date(2024, 4, 5, 11, 30, 0, 0, time.UTC); return &t }(),
			ExpiresAt:       time.Date(2024, 4, 30, 23, 59, 59, 0, time.UTC),
		},
	}

	for i := range yieldClaims {
		if err := db.Create(&yieldClaims[i]).Error; err != nil {
			return err
		}
	}
	log.Printf("Created %d yield claims", len(yieldClaims))

	// Create sample redemption requests
	redemptions := []models.Redemption{
		{
			SukukSeriesID:    sukukSeries[0].ID, // PLN
			InvestmentID:     investments[1].ID,
			InvestorAddress:  "0xbcd2345678901234567890123456789012345678",
			TokenAmount:      "5000000000000000000000000", // 5M tokens
			RedemptionAmount: "5000000000000000000000000", // 5M IDRX
			Status:           "requested",
			RequestReason:    "Need liquidity for other investments",
			RequestedAt:      time.Date(2024, 4, 10, 16, 45, 0, 0, time.UTC),
		},
		{
			SukukSeriesID:    sukukSeries[1].ID, // Antam
			InvestmentID:     investments[2].ID,
			InvestorAddress:  "0xabc1234567890123456789012345678901234567",
			TokenAmount:      "7500000000000000000000000", // 7.5M tokens
			RedemptionAmount: "7500000000000000000000000", // 7.5M IDRX
			Status:           "approved",
			RequestReason:    "Portfolio rebalancing",
			ApprovalNotes:    "Approved for partial redemption",
			TransactionHash:  "0x2223456789012345678901234567890123456789012345678901234567890123",
			RequestedAt:      time.Date(2024, 3, 20, 10, 0, 0, 0, time.UTC),
			ApprovedAt:       func() *time.Time { t := time.Date(2024, 3, 22, 14, 30, 0, 0, time.UTC); return &t }(),
		},
	}

	for i := range redemptions {
		if err := db.Create(&redemptions[i]).Error; err != nil {
			return err
		}
	}
	log.Printf("Created %d redemptions", len(redemptions))

	log.Println("Database seeding completed successfully!")
	return nil
}
