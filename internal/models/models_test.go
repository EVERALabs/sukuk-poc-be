package models

import (
	"testing"
	"time"

	"github.com/kadzu/sukuk-poc-be/internal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type ModelsTestSuite struct {
	suite.Suite
	db *gorm.DB
}

func (suite *ModelsTestSuite) SetupSuite() {
	// Initialize test logger (silent)
	logger.Init("error", "json")

	// Create test database connection
	dsn := "host=localhost user=postgres password=postgres dbname=sukuk_poc_test port=5432 sslmode=disable TimeZone=UTC"
	
	// Create test database if it doesn't exist
	suite.createTestDatabase()
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	assert.NoError(suite.T(), err, "Failed to connect to test database")

	// Run migrations
	err = db.AutoMigrate(AllModels()...)
	assert.NoError(suite.T(), err, "Failed to run test migrations")

	suite.db = db
}

func (suite *ModelsTestSuite) createTestDatabase() {
	adminDSN := "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable TimeZone=UTC"
	adminDB, err := gorm.Open(postgres.Open(adminDSN), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	if err != nil {
		return // Skip if can't connect
	}

	var exists bool
	adminDB.Raw("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = ?)", "sukuk_poc_test").Scan(&exists)
	if !exists {
		adminDB.Exec("CREATE DATABASE sukuk_poc_test")
	}

	sqlDB, _ := adminDB.DB()
	sqlDB.Close()
}

func (suite *ModelsTestSuite) TearDownSuite() {
	if suite.db != nil {
		sqlDB, _ := suite.db.DB()
		sqlDB.Close()
	}
}

func (suite *ModelsTestSuite) SetupTest() {
	// Clean database before each test
	tables := []string{"events", "redemptions", "yield_claims", "investments", "sukuk_series", "companies"}
	for _, table := range tables {
		suite.db.Exec("TRUNCATE TABLE " + table + " RESTART IDENTITY CASCADE")
	}
}

func (suite *ModelsTestSuite) TestCompanyModel() {
	company := &Company{
		Name:          "Test Company",
		Description:   "A test company",
		Website:       "https://test.com",
		Industry:      "Technology",
		Email:         "test@company.com",
		WalletAddress: "0x1234567890123456789012345678901234567890",
		IsActive:      true,
	}

	// Test creation
	err := suite.db.Create(company).Error
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), company.ID)
	assert.NotZero(suite.T(), company.CreatedAt)
	assert.NotZero(suite.T(), company.UpdatedAt)

	// Test retrieval
	var retrieved Company
	err = suite.db.First(&retrieved, company.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), company.Name, retrieved.Name)
	assert.Equal(suite.T(), company.Email, retrieved.Email)
	assert.Equal(suite.T(), company.WalletAddress, retrieved.WalletAddress)

	// Test update
	company.Name = "Updated Company"
	err = suite.db.Save(company).Error
	assert.NoError(suite.T(), err)

	err = suite.db.First(&retrieved, company.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Company", retrieved.Name)
}

func (suite *ModelsTestSuite) TestSukukSeriesModel() {
	// Create company first
	company := &Company{
		Name:          "Test Company",
		Description:   "A test company",
		Website:       "https://test.com",
		Industry:      "Technology",
		Email:         "test@company.com",
		WalletAddress: "0x1234567890123456789012345678901234567890",
		IsActive:      true,
	}
	err := suite.db.Create(company).Error
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), company.ID)

	sukukSeries := &SukukSeries{
		CompanyID:         company.ID,
		Name:              "Test Sukuk",
		Symbol:            "TST",
		Description:       "Test sukuk series",
		TokenAddress:      "0x1234567890123456789012345678901234567891",
		TotalSupply:       "1000000000000000000000000",
		OutstandingSupply: "500000000000000000000000",
		YieldRate:         8.5,
		MaturityDate:      time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
		PaymentFrequency:  4,
		MinInvestment:     "1000000000000000000000",
		MaxInvestment:     "10000000000000000000000",
		Status:            SukukStatusActive,
		IsRedeemable:      true,
	}

	// Test creation
	err = suite.db.Create(sukukSeries).Error
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), sukukSeries.ID)

	// Test relationship loading
	var retrieved SukukSeries
	err = suite.db.Preload("Company").First(&retrieved, sukukSeries.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), company.Name, retrieved.Company.Name)
	assert.Equal(suite.T(), sukukSeries.Symbol, retrieved.Symbol)
	assert.Equal(suite.T(), sukukSeries.YieldRate, retrieved.YieldRate)
}

func (suite *ModelsTestSuite) TestInvestmentModel() {
	// Create company and sukuk series
	company := &Company{
		Name:          "Test Company",
		Description:   "A test company",
		Website:       "https://test.com",
		Industry:      "Technology",
		Email:         "test@company.com",
		WalletAddress: "0x1234567890123456789012345678901234567890",
		IsActive:      true,
	}
	err := suite.db.Create(company).Error
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), company.ID)
	
	sukukSeries := &SukukSeries{
		CompanyID:         company.ID,
		Name:              "Test Sukuk",
		Symbol:            "TST",
		Description:       "Test sukuk series",
		TokenAddress:      "0x1234567890123456789012345678901234567891",
		TotalSupply:       "1000000000000000000000000",
		OutstandingSupply: "500000000000000000000000",
		YieldRate:         8.5,
		MaturityDate:      time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
		PaymentFrequency:  4,
		MinInvestment:     "1000000000000000000000",
		MaxInvestment:     "10000000000000000000000",
		Status:            SukukStatusActive,
		IsRedeemable:      true,
	}
	err = suite.db.Create(sukukSeries).Error
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), sukukSeries.ID)

	investment := &Investment{
		SukukSeriesID:   sukukSeries.ID,
		InvestorAddress: "0xabc1234567890123456789012345678901234567",
		InvestorEmail:   "investor@test.com",
		Amount:          "5000000000000000000000",
		TokenAmount:     "5000000000000000000000",
		Status:          InvestmentStatusActive,
		TransactionHash: "0xdef1234567890123456789012345678901234567890123456789012345678901",
		BlockNumber:     123456,
	}

	// Test creation
	err = suite.db.Create(investment).Error
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), investment.ID)

	// Test relationship loading
	var retrieved Investment
	err = suite.db.Preload("SukukSeries").Preload("SukukSeries.Company").First(&retrieved, investment.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), sukukSeries.Name, retrieved.SukukSeries.Name)
	assert.Equal(suite.T(), company.Name, retrieved.SukukSeries.Company.Name)
	assert.Equal(suite.T(), investment.InvestorAddress, retrieved.InvestorAddress)
}

func (suite *ModelsTestSuite) TestYieldClaimModel() {
	// Create test data
	company := &Company{
		Name:          "Test Company",
		Description:   "A test company",
		Website:       "https://test.com",
		Industry:      "Technology",
		Email:         "test@company.com",
		WalletAddress: "0x1234567890123456789012345678901234567890",
		IsActive:      true,
	}
	err := suite.db.Create(company).Error
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), company.ID)
	
	sukukSeries := &SukukSeries{
		CompanyID:         company.ID,
		Name:              "Test Sukuk",
		Symbol:            "TST",
		Description:       "Test sukuk series",
		TokenAddress:      "0x1234567890123456789012345678901234567891",
		TotalSupply:       "1000000000000000000000000",
		OutstandingSupply: "500000000000000000000000",
		YieldRate:         8.5,
		MaturityDate:      time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
		PaymentFrequency:  4,
		MinInvestment:     "1000000000000000000000",
		MaxInvestment:     "10000000000000000000000",
		Status:            SukukStatusActive,
		IsRedeemable:      true,
	}
	err = suite.db.Create(sukukSeries).Error
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), sukukSeries.ID)
	
	investment := &Investment{
		SukukSeriesID:   sukukSeries.ID,
		InvestorAddress: "0xabc1234567890123456789012345678901234567",
		InvestorEmail:   "investor@test.com",
		Amount:          "5000000000000000000000",
		TokenAmount:     "5000000000000000000000",
		Status:          InvestmentStatusActive,
		TransactionHash: "0xdef1234567890123456789012345678901234567890123456789012345678901",
		BlockNumber:     123456,
	}
	err = suite.db.Create(investment).Error
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), investment.ID)

	yieldClaim := &YieldClaim{
		SukukSeriesID:   sukukSeries.ID,
		InvestmentID:    investment.ID,
		InvestorAddress: investment.InvestorAddress,
		YieldAmount:     "250000000000000000000",
		PeriodStart:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:       time.Date(2024, 3, 31, 23, 59, 59, 0, time.UTC),
		Status:          YieldClaimStatusPending,
		ExpiresAt:       time.Date(2024, 4, 30, 23, 59, 59, 0, time.UTC),
	}

	// Test creation
	err = suite.db.Create(yieldClaim).Error
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), yieldClaim.ID)

	// Test relationship loading
	var retrieved YieldClaim
	err = suite.db.Preload("SukukSeries").Preload("Investment").First(&retrieved, yieldClaim.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), sukukSeries.Name, retrieved.SukukSeries.Name)
	assert.Equal(suite.T(), investment.Amount, retrieved.Investment.Amount)
	assert.Equal(suite.T(), yieldClaim.YieldAmount, retrieved.YieldAmount)
}

func (suite *ModelsTestSuite) TestRedemptionModel() {
	// Create test data
	company := &Company{
		Name:          "Test Company",
		Description:   "A test company",
		Website:       "https://test.com",
		Industry:      "Technology",
		Email:         "test@company.com",
		WalletAddress: "0x1234567890123456789012345678901234567890",
		IsActive:      true,
	}
	err := suite.db.Create(company).Error
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), company.ID)
	
	sukukSeries := &SukukSeries{
		CompanyID:         company.ID,
		Name:              "Test Sukuk",
		Symbol:            "TST",
		Description:       "Test sukuk series",
		TokenAddress:      "0x1234567890123456789012345678901234567891",
		TotalSupply:       "1000000000000000000000000",
		OutstandingSupply: "500000000000000000000000",
		YieldRate:         8.5,
		MaturityDate:      time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
		PaymentFrequency:  4,
		MinInvestment:     "1000000000000000000000",
		MaxInvestment:     "10000000000000000000000",
		Status:            SukukStatusActive,
		IsRedeemable:      true,
	}
	err = suite.db.Create(sukukSeries).Error
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), sukukSeries.ID)
	
	investment := &Investment{
		SukukSeriesID:   sukukSeries.ID,
		InvestorAddress: "0xabc1234567890123456789012345678901234567",
		InvestorEmail:   "investor@test.com",
		Amount:          "5000000000000000000000",
		TokenAmount:     "5000000000000000000000",
		Status:          InvestmentStatusActive,
		TransactionHash: "0xdef1234567890123456789012345678901234567890123456789012345678901",
		BlockNumber:     123456,
	}
	err = suite.db.Create(investment).Error
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), investment.ID)

	redemption := &Redemption{
		SukukSeriesID:    sukukSeries.ID,
		InvestmentID:     investment.ID,
		InvestorAddress:  investment.InvestorAddress,
		TokenAmount:      "2000000000000000000000",
		RedemptionAmount: "2000000000000000000000",
		Status:           RedemptionStatusRequested,
		RequestReason:    "Need liquidity",
		RequestedAt:      time.Now(),
	}

	// Test creation
	err = suite.db.Create(redemption).Error
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), redemption.ID)

	// Test relationship loading
	var retrieved Redemption
	err = suite.db.Preload("SukukSeries").Preload("Investment").First(&retrieved, redemption.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), sukukSeries.Name, retrieved.SukukSeries.Name)
	assert.Equal(suite.T(), investment.Amount, retrieved.Investment.Amount)
	assert.Equal(suite.T(), redemption.TokenAmount, retrieved.TokenAmount)
}

func (suite *ModelsTestSuite) TestEventModel() {
	event := &Event{
		EventName:       "Investment",
		BlockNumber:     123456,
		TxHash:          "0x1234567890123456789012345678901234567890123456789012345678901234",
		ContractAddress: "0x1234567890123456789012345678901234567890",
		Data: JSON{
			"investor":        "0xabc1234567890123456789012345678901234567",
			"amount":          "1000000000000000000000",
			"sukuk_series_id": 1,
		},
		ChainID:   84532,
		Processed: false,
	}

	// Test creation
	err := suite.db.Create(event).Error
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), event.ID)

	// Test retrieval
	var retrieved Event
	err = suite.db.First(&retrieved, event.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), event.EventName, retrieved.EventName)
	assert.Equal(suite.T(), event.TxHash, retrieved.TxHash)
	assert.Equal(suite.T(), event.ChainID, retrieved.ChainID)
	assert.False(suite.T(), retrieved.Processed)

	// Test JSON data
	assert.NotNil(suite.T(), retrieved.Data)
	assert.Equal(suite.T(), "0xabc1234567890123456789012345678901234567", retrieved.Data["investor"])
}

func (suite *ModelsTestSuite) TestBeforeCreateHooks() {
	event := &Event{
		EventName:       "Test",
		BlockNumber:     123,
		TxHash:          "0xABCDEF1234567890123456789012345678901234567890123456789012345678",
		ContractAddress: "0xABCDEF1234567890123456789012345678901234",
		Data:            JSON{},
		ChainID:         84532,
	}

	err := suite.db.Create(event).Error
	assert.NoError(suite.T(), err)

	// Verify addresses were normalized to lowercase
	var retrieved Event
	err = suite.db.First(&retrieved, event.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "0xabcdef1234567890123456789012345678901234567890123456789012345678", retrieved.TxHash)
	assert.Equal(suite.T(), "0xabcdef1234567890123456789012345678901234", retrieved.ContractAddress)
}

func (suite *ModelsTestSuite) TestAllModelsFunction() {
	models := AllModels()
	assert.Len(suite.T(), models, 6)

	// Verify all expected models are included
	modelTypes := make(map[string]bool)
	for _, model := range models {
		switch model.(type) {
		case *Company:
			modelTypes["Company"] = true
		case *SukukSeries:
			modelTypes["SukukSeries"] = true
		case *Investment:
			modelTypes["Investment"] = true
		case *YieldClaim:
			modelTypes["YieldClaim"] = true
		case *Redemption:
			modelTypes["Redemption"] = true
		case *Event:
			modelTypes["Event"] = true
		}
	}

	assert.True(suite.T(), modelTypes["Company"])
	assert.True(suite.T(), modelTypes["SukukSeries"])
	assert.True(suite.T(), modelTypes["Investment"])
	assert.True(suite.T(), modelTypes["YieldClaim"])
	assert.True(suite.T(), modelTypes["Redemption"])
	assert.True(suite.T(), modelTypes["Event"])
}

func (suite *ModelsTestSuite) TestModelConstants() {
	// Test SukukStatus constants
	assert.Equal(suite.T(), "active", string(SukukStatusActive))
	assert.Equal(suite.T(), "matured", string(SukukStatusMatured))
	assert.Equal(suite.T(), "suspended", string(SukukStatusSuspended))

	// Test InvestmentStatus constants
	assert.Equal(suite.T(), "active", string(InvestmentStatusActive))
	assert.Equal(suite.T(), "redeemed", string(InvestmentStatusRedeemed))
	assert.Equal(suite.T(), "matured", string(InvestmentStatusMatured))

	// Test YieldClaimStatus constants
	assert.Equal(suite.T(), "pending", string(YieldClaimStatusPending))
	assert.Equal(suite.T(), "claimed", string(YieldClaimStatusClaimed))
	assert.Equal(suite.T(), "expired", string(YieldClaimStatusExpired))

	// Test RedemptionStatus constants
	assert.Equal(suite.T(), "requested", string(RedemptionStatusRequested))
	assert.Equal(suite.T(), "approved", string(RedemptionStatusApproved))
	assert.Equal(suite.T(), "rejected", string(RedemptionStatusRejected))
	assert.Equal(suite.T(), "completed", string(RedemptionStatusCompleted))
	assert.Equal(suite.T(), "cancelled", string(RedemptionStatusCancelled))
}

// Run the test suite
func TestModelsTestSuite(t *testing.T) {
	suite.Run(t, new(ModelsTestSuite))
}