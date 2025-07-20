package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kadzu/sukuk-poc-be/internal/models"
	"github.com/kadzu/sukuk-poc-be/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CompanyTestSuite struct {
	suite.Suite
	testCfg *testutil.TestConfig
	router  *gin.Engine
}

func (suite *CompanyTestSuite) SetupSuite() {
	suite.testCfg = testutil.SetupTestEnvironment(suite.T())
	suite.router = gin.New()
	
	// Setup routes
	suite.router.GET("/api/v1/companies", ListCompanies)
	suite.router.GET("/api/v1/companies/:id", GetCompany)
	suite.router.GET("/api/v1/companies/:id/sukuks", GetCompanySukuks)
}

func (suite *CompanyTestSuite) TearDownSuite() {
	testutil.CleanupTestEnvironment(suite.T(), suite.testCfg)
}

func (suite *CompanyTestSuite) SetupTest() {
	// Clean database before each test
	suite.testCfg.DB.Exec("TRUNCATE TABLE companies RESTART IDENTITY CASCADE")
}

func (suite *CompanyTestSuite) TestListCompanies_Empty() {
	// Create test request
	req := testutil.MakeTestRequest("GET", "/api/v1/companies", nil, nil)
	w := httptest.NewRecorder()

	// Execute request
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), float64(0), response["count"])
	assert.Empty(suite.T(), response["data"])
}

func (suite *CompanyTestSuite) TestListCompanies_WithData() {
	// Create test companies
	company1 := testutil.CreateTestCompany(suite.testCfg.DB)
	company2 := &models.Company{
		Name:          "Second Company",
		Description:   "Another test company",
		Website:       "https://second.com",
		Industry:      "Finance",
		Email:         "info@second.com",
		WalletAddress: "0x2222222222222222222222222222222222222222",
		IsActive:      true,
	}
	suite.testCfg.DB.Create(company2)

	// Create test request
	req := testutil.MakeTestRequest("GET", "/api/v1/companies", nil, nil)
	w := httptest.NewRecorder()

	// Execute request
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), float64(2), response["count"])
	
	companies := response["data"].([]interface{})
	assert.Len(suite.T(), companies, 2)

	// Check first company
	firstCompany := companies[0].(map[string]interface{})
	assert.Equal(suite.T(), company1.Name, firstCompany["name"])
	assert.Equal(suite.T(), company1.Email, firstCompany["email"])
	assert.Equal(suite.T(), company1.WalletAddress, firstCompany["wallet_address"])
}

func (suite *CompanyTestSuite) TestListCompanies_OnlyActiveCompanies() {
	// Create active and inactive companies
	activeCompany := testutil.CreateTestCompany(suite.testCfg.DB)
	inactiveCompany := &models.Company{
		Name:          "Inactive Company",
		Description:   "An inactive company",
		Website:       "https://inactive.com",
		Industry:      "Mining",
		Email:         "info@inactive.com",
		WalletAddress: "0x3333333333333333333333333333333333333333",
		IsActive:      true, // Create as active first
	}
	suite.testCfg.DB.Create(inactiveCompany)
	
	// Then update to inactive to bypass GORM default value issue
	suite.testCfg.DB.Model(inactiveCompany).Update("is_active", false)

	// ListCompanies should only return active companies by default
	req := testutil.MakeTestRequest("GET", "/api/v1/companies", nil, nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// Should only return the active company
	assert.Equal(suite.T(), float64(1), response["count"])
	
	companies := response["data"].([]interface{})
	assert.Len(suite.T(), companies, 1)
	
	company := companies[0].(map[string]interface{})
	assert.Equal(suite.T(), activeCompany.Name, company["name"])
	assert.Equal(suite.T(), true, company["is_active"])
}

func (suite *CompanyTestSuite) TestGetCompany_Success() {
	// Create test company
	company := testutil.CreateTestCompany(suite.testCfg.DB)

	// Create test request
	req := testutil.MakeTestRequest("GET", "/api/v1/companies/1", nil, nil)
	w := httptest.NewRecorder()

	// Execute request
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	companyData := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), company.Name, companyData["name"])
	assert.Equal(suite.T(), company.Email, companyData["email"])
	assert.Equal(suite.T(), company.WalletAddress, companyData["wallet_address"])
	assert.Equal(suite.T(), company.Industry, companyData["industry"])
}

func (suite *CompanyTestSuite) TestGetCompany_NotFound() {
	// Create test request for non-existent company
	req := testutil.MakeTestRequest("GET", "/api/v1/companies/999", nil, nil)
	w := httptest.NewRecorder()

	// Execute request
	suite.router.ServeHTTP(w, req)

	// Assert response
	testutil.AssertErrorResponse(suite.T(), w, http.StatusNotFound, "Company not found")
}

func (suite *CompanyTestSuite) TestGetCompany_InvalidID() {
	// Create test request with invalid ID
	req := testutil.MakeTestRequest("GET", "/api/v1/companies/invalid", nil, nil)
	w := httptest.NewRecorder()

	// Execute request
	suite.router.ServeHTTP(w, req)

	// Assert response
	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "Invalid company ID")
}

func (suite *CompanyTestSuite) TestGetCompanySukuks_Success() {
	// Create test company
	company := testutil.CreateTestCompany(suite.testCfg.DB)
	
	// Create test sukuk series
	sukuk1 := testutil.CreateTestSukukSeries(suite.testCfg.DB, company.ID)
	sukuk2 := &models.SukukSeries{
		CompanyID:         company.ID,
		Name:              "Second Sukuk",
		Symbol:            "SEC",
		Description:       "Second test sukuk",
		TokenAddress:      "0x2222222222222222222222222222222222222222",
		TotalSupply:       "2000000000000000000000000",
		OutstandingSupply: "1000000000000000000000000",
		YieldRate:         7.5,
		PaymentFrequency:  4,
		MinInvestment:     "2000000000000000000000",
		MaxInvestment:     "20000000000000000000000",
		Status:            models.SukukStatusActive,
		IsRedeemable:      true,
	}
	suite.testCfg.DB.Create(sukuk2)

	// Create test request
	req := testutil.MakeTestRequest("GET", "/api/v1/companies/1/sukuks", nil, nil)
	w := httptest.NewRecorder()

	// Execute request
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), float64(2), response["count"])
	
	sukuks := response["data"].([]interface{})
	assert.Len(suite.T(), sukuks, 2)

	// Check first sukuk
	firstSukuk := sukuks[0].(map[string]interface{})
	assert.Equal(suite.T(), sukuk1.Name, firstSukuk["name"])
	assert.Equal(suite.T(), sukuk1.Symbol, firstSukuk["symbol"])
}

func (suite *CompanyTestSuite) TestGetCompanySukuks_EmptyList() {
	// Create test company without sukuks
	testutil.CreateTestCompany(suite.testCfg.DB)

	// Create test request
	req := testutil.MakeTestRequest("GET", "/api/v1/companies/1/sukuks", nil, nil)
	w := httptest.NewRecorder()

	// Execute request
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), float64(0), response["count"])
	assert.Empty(suite.T(), response["data"])
}

func (suite *CompanyTestSuite) TestGetCompanySukuks_CompanyNotFound() {
	// Create test request for non-existent company
	req := testutil.MakeTestRequest("GET", "/api/v1/companies/999/sukuks", nil, nil)
	w := httptest.NewRecorder()

	// Execute request
	suite.router.ServeHTTP(w, req)

	// Assert response - should return empty list, not error
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), float64(0), response["count"])
	assert.Empty(suite.T(), response["data"])
}

// Run the test suite
func TestCompanyTestSuite(t *testing.T) {
	suite.Run(t, new(CompanyTestSuite))
}