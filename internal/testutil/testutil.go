package testutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"sukuk-be/internal/config"
	"sukuk-be/internal/database"
	"sukuk-be/internal/logger"
	"sukuk-be/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// TestConfig holds test configuration
type TestConfig struct {
	Config *config.Config
	DB     *gorm.DB
}

// SetupTestEnvironment initializes test environment with test database
func SetupTestEnvironment(t *testing.T) *TestConfig {
	// Set gin to test mode
	gin.SetMode(gin.TestMode)

	// Initialize test logger (silent)
	logger.Init("error", "json")

	// Create test configuration
	cfg := &config.Config{
		App: config.AppConfig{
			Name:        "sukuk-poc-api-test",
			Environment: "test",
			Port:        8081,
			Debug:       false,
		},
		Database: config.DatabaseConfig{
			Host:            "localhost",
			Port:            5432,
			User:            "postgres",
			Password:        "postgres",
			DBName:          "sukuk_poc_test",
			SSLMode:         "disable",
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 300000000000, // 5 minutes in nanoseconds
		},
		API: config.APIConfig{
			APIKey:          "test-api-key",
			RateLimitPerMin: 1000,
			AllowedOrigins:  []string{"*"},
		},
	}

	// Setup test database
	db := setupTestDatabase(t, cfg)

	return &TestConfig{
		Config: cfg,
		DB:     db,
	}
}

// setupTestDatabase creates and configures test database
func setupTestDatabase(t *testing.T, cfg *config.Config) *gorm.DB {
	// Create test database if it doesn't exist
	createTestDatabase(t, cfg)

	// Connect to test database
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=UTC",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	assert.NoError(t, err, "Failed to connect to test database")

	// Run migrations
	err = db.AutoMigrate(models.AllModels()...)
	assert.NoError(t, err, "Failed to run test migrations")

	// Set global database for handlers
	database.DB = db

	return db
}

// createTestDatabase creates test database if it doesn't exist
func createTestDatabase(t *testing.T, cfg *config.Config) {
	// Connect to postgres database to create test database
	adminDSN := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=postgres port=%d sslmode=%s TimeZone=UTC",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)

	adminDB, err := gorm.Open(postgres.Open(adminDSN), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Silent),
	})
	assert.NoError(t, err, "Failed to connect to postgres for test database creation")

	// Check if test database exists
	var exists bool
	err = adminDB.Raw("SELECT EXISTS(SELECT datname FROM pg_catalog.pg_database WHERE datname = ?)", cfg.Database.DBName).Scan(&exists).Error
	assert.NoError(t, err, "Failed to check if test database exists")

	if !exists {
		// Create test database
		createSQL := fmt.Sprintf("CREATE DATABASE %s", cfg.Database.DBName)
		err = adminDB.Exec(createSQL).Error
		assert.NoError(t, err, "Failed to create test database")
	}

	// Close admin connection
	sqlDB, _ := adminDB.DB()
	sqlDB.Close()
}

// CleanupTestEnvironment cleans up test environment
func CleanupTestEnvironment(t *testing.T, testCfg *TestConfig) {
	if testCfg.DB != nil {
		// Clean all tables
		tables := []string{"events", "redemptions", "yield_claims", "investments", "sukuk_series", "companies"}
		for _, table := range tables {
			testCfg.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
		}

		// Close database connection
		sqlDB, _ := testCfg.DB.DB()
		sqlDB.Close()
	}
}

// CreateTestCompany creates a test company
func CreateTestCompany(db *gorm.DB) *models.Company {
	company := &models.Company{
		Name:          "Test Company",
		Description:   "A test company for testing",
		Website:       "https://test.com",
		Industry:      "Testing",
		Email:         "test@company.com",
		WalletAddress: "0x1234567890123456789012345678901234567890",
		IsActive:      true,
	}
	db.Create(company)
	return company
}

// CreateTestSukukSeries creates a test sukuk series
func CreateTestSukukSeries(db *gorm.DB, companyID uint) *models.SukukSeries {
	sukuk := &models.SukukSeries{
		CompanyID:         companyID,
		Name:              "Test Sukuk Series",
		Symbol:            "TEST",
		Description:       "Test sukuk for testing",
		TokenAddress:      "0x1234567890123456789012345678901234567891",
		TotalSupply:       "1000000000000000000000000",
		OutstandingSupply: "500000000000000000000000",
		YieldRate:         8.5,
		PaymentFrequency:  4,
		MinInvestment:     "1000000000000000000000",
		MaxInvestment:     "10000000000000000000000",
		Status:            models.SukukStatusActive,
		IsRedeemable:      true,
	}
	db.Create(sukuk)
	return sukuk
}

// CreateTestInvestment creates a test investment
func CreateTestInvestment(db *gorm.DB, sukukID uint) *models.Investment {
	investment := &models.Investment{
		SukukSeriesID:   sukukID,
		InvestorAddress: "0xabc1234567890123456789012345678901234567",
		InvestorEmail:   "investor@test.com",
		Amount:          "5000000000000000000000",
		TokenAmount:     "5000000000000000000000",
		Status:          models.InvestmentStatusActive,
		TransactionHash: "0xdef1234567890123456789012345678901234567890123456789012345678901",
		BlockNumber:     123456,
	}
	db.Create(investment)
	return investment
}

// MakeTestRequest creates a test HTTP request
func MakeTestRequest(method, url string, body interface{}, headers map[string]string) *http.Request {
	var reqBody io.Reader

	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, _ := http.NewRequest(method, url, reqBody)

	// Set default headers
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Set custom headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return req
}

// MakeTestMultipartRequest creates a test multipart form request for file uploads
func MakeTestMultipartRequest(url string, params map[string]string, filename, fieldname string, fileContent []byte) *http.Request {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add form fields
	for key, val := range params {
		writer.WriteField(key, val)
	}

	// Add file
	part, _ := writer.CreateFormFile(fieldname, filename)
	part.Write(fileContent)
	writer.Close()

	req, _ := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req
}

// CreateTestFile creates a temporary test file
func CreateTestFile(t *testing.T, filename, content string) string {
	tempDir := t.TempDir()
	filePath := filepath.Join(tempDir, filename)

	err := os.WriteFile(filePath, []byte(content), 0644)
	assert.NoError(t, err, "Failed to create test file")

	return filePath
}

// AssertJSONResponse asserts that response contains expected JSON
func AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedData interface{}) {
	assert.Equal(t, expectedStatus, w.Code)
	assert.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	if expectedData != nil {
		var actualData interface{}
		err := json.Unmarshal(w.Body.Bytes(), &actualData)
		assert.NoError(t, err, "Response should be valid JSON")

		expectedJSON, _ := json.Marshal(expectedData)
		actualJSON, _ := json.Marshal(actualData)
		assert.JSONEq(t, string(expectedJSON), string(actualJSON))
	}
}

// AssertErrorResponse asserts that response contains an error
func AssertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedErrorMessage string) {
	assert.Equal(t, expectedStatus, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err, "Response should be valid JSON")

	errorMsg, exists := response["error"]
	assert.True(t, exists, "Response should contain error field")

	if expectedErrorMessage != "" {
		assert.Contains(t, errorMsg.(string), expectedErrorMessage)
	}
}

// AuthHeaders returns headers with API key for authenticated requests
func AuthHeaders(apiKey string) map[string]string {
	return map[string]string{
		"X-API-Key": apiKey,
	}
}

// SeedTestData populates database with test data
func SeedTestData(db *gorm.DB) {
	// Create test companies
	companies := []models.Company{
		{
			Name:          "Test Corp",
			Description:   "Test corporation",
			Website:       "https://testcorp.com",
			Industry:      "Technology",
			Email:         "info@testcorp.com",
			WalletAddress: "0x1111111111111111111111111111111111111111",
			IsActive:      true,
		},
		{
			Name:          "Demo Inc",
			Description:   "Demo company",
			Website:       "https://demo.com",
			Industry:      "Finance",
			Email:         "contact@demo.com",
			WalletAddress: "0x2222222222222222222222222222222222222222",
			IsActive:      true,
		},
	}

	for i := range companies {
		db.Create(&companies[i])
	}
}
