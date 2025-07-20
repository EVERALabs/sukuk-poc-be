package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/kadzu/sukuk-poc-be/internal/models"
	"github.com/kadzu/sukuk-poc-be/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type CompanyManagementTestSuite struct {
	suite.Suite
	testCfg *testutil.TestConfig
	router  *gin.Engine
}

func (suite *CompanyManagementTestSuite) SetupSuite() {
	suite.testCfg = testutil.SetupTestEnvironment(suite.T())
	suite.router = gin.New()
	
	// Setup routes
	suite.router.POST("/api/v1/admin/companies", CreateCompany)
	suite.router.PUT("/api/v1/admin/companies/:id", UpdateCompany)
	suite.router.POST("/api/v1/admin/companies/:id/upload-logo", UploadCompanyLogo)
}

func (suite *CompanyManagementTestSuite) TearDownSuite() {
	testutil.CleanupTestEnvironment(suite.T(), suite.testCfg)
}

func (suite *CompanyManagementTestSuite) SetupTest() {
	// Clean database before each test
	suite.testCfg.DB.Exec("TRUNCATE TABLE companies RESTART IDENTITY CASCADE")
	
	// Create test uploads directory
	os.MkdirAll("./uploads/logos", 0755)
}

func (suite *CompanyManagementTestSuite) TearDownTest() {
	// Clean up test files
	os.RemoveAll("./uploads")
}

func (suite *CompanyManagementTestSuite) TestCreateCompany_Success() {
	requestBody := CreateCompanyRequest{
		Name:          "New Test Company",
		Description:   "A new test company",
		Website:       "https://newtest.com",
		Industry:      "Technology",
		Email:         "info@newtest.com",
		WalletAddress: "0x1234567890123456789012345678901234567890",
	}

	// Create test request
	req := testutil.MakeTestRequest("POST", "/api/v1/admin/companies", requestBody, nil)
	w := httptest.NewRecorder()

	// Execute request
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Company created successfully", response["message"])
	
	companyData := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), requestBody.Name, companyData["name"])
	assert.Equal(suite.T(), requestBody.Email, companyData["email"])
	assert.Equal(suite.T(), requestBody.WalletAddress, companyData["wallet_address"])
	assert.Equal(suite.T(), true, companyData["is_active"])

	// Verify company was created in database
	var company models.Company
	err = suite.testCfg.DB.First(&company, "email = ?", requestBody.Email).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), requestBody.Name, company.Name)
}

func (suite *CompanyManagementTestSuite) TestCreateCompany_ValidationError() {
	requestBody := CreateCompanyRequest{
		Name:        "Test Company",
		Description: "Missing required fields",
		// Missing Email and WalletAddress
	}

	// Create test request
	req := testutil.MakeTestRequest("POST", "/api/v1/admin/companies", requestBody, nil)
	w := httptest.NewRecorder()

	// Execute request
	suite.router.ServeHTTP(w, req)

	// Assert response
	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "")
}

func (suite *CompanyManagementTestSuite) TestCreateCompany_InvalidEmail() {
	requestBody := CreateCompanyRequest{
		Name:          "Test Company",
		Email:         "invalid-email",
		WalletAddress: "0x1234567890123456789012345678901234567890",
	}

	// Create test request
	req := testutil.MakeTestRequest("POST", "/api/v1/admin/companies", requestBody, nil)
	w := httptest.NewRecorder()

	// Execute request
	suite.router.ServeHTTP(w, req)

	// Assert response
	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "")
}

func (suite *CompanyManagementTestSuite) TestUpdateCompany_Success() {
	// Create test company
	company := testutil.CreateTestCompany(suite.testCfg.DB)

	requestBody := UpdateCompanyRequest{
		Name:        "Updated Company Name",
		Description: "Updated description",
		Website:     "https://updated.com",
		Industry:    "Updated Industry",
	}

	// Create test request
	req := testutil.MakeTestRequest("PUT", fmt.Sprintf("/api/v1/admin/companies/%d", company.ID), requestBody, nil)
	w := httptest.NewRecorder()

	// Execute request
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Company updated successfully", response["message"])
	
	companyData := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), requestBody.Name, companyData["name"])
	assert.Equal(suite.T(), requestBody.Description, companyData["description"])
	assert.Equal(suite.T(), requestBody.Website, companyData["website"])

	// Verify company was updated in database
	var updatedCompany models.Company
	err = suite.testCfg.DB.First(&updatedCompany, company.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), requestBody.Name, updatedCompany.Name)
	assert.Equal(suite.T(), requestBody.Description, updatedCompany.Description)
}

func (suite *CompanyManagementTestSuite) TestUpdateCompany_PartialUpdate() {
	// Create test company
	company := testutil.CreateTestCompany(suite.testCfg.DB)
	originalName := company.Name

	requestBody := UpdateCompanyRequest{
		Description: "Only updating description",
	}

	// Create test request
	req := testutil.MakeTestRequest("PUT", fmt.Sprintf("/api/v1/admin/companies/%d", company.ID), requestBody, nil)
	w := httptest.NewRecorder()

	// Execute request
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Verify only description was updated
	var updatedCompany models.Company
	err := suite.testCfg.DB.First(&updatedCompany, company.ID).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), originalName, updatedCompany.Name) // Name unchanged
	assert.Equal(suite.T(), requestBody.Description, updatedCompany.Description) // Description updated
}

func (suite *CompanyManagementTestSuite) TestUpdateCompany_NotFound() {
	requestBody := UpdateCompanyRequest{
		Name: "Updated Name",
	}

	// Create test request for non-existent company
	req := testutil.MakeTestRequest("PUT", "/api/v1/admin/companies/999", requestBody, nil)
	w := httptest.NewRecorder()

	// Execute request
	suite.router.ServeHTTP(w, req)

	// Assert response
	testutil.AssertErrorResponse(suite.T(), w, http.StatusNotFound, "Company not found")
}

func (suite *CompanyManagementTestSuite) TestUpdateCompany_InvalidID() {
	requestBody := UpdateCompanyRequest{
		Name: "Updated Name",
	}

	// Create test request with invalid ID
	req := testutil.MakeTestRequest("PUT", "/api/v1/admin/companies/invalid", requestBody, nil)
	w := httptest.NewRecorder()

	// Execute request
	suite.router.ServeHTTP(w, req)

	// Assert response
	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "Invalid company ID")
}

func (suite *CompanyManagementTestSuite) TestUploadCompanyLogo_Success() {
	// Create test company
	company := testutil.CreateTestCompany(suite.testCfg.DB)

	// Create test image content
	imageContent := []byte("fake png content")

	// Create multipart request
	req := testutil.MakeTestMultipartRequest(
		fmt.Sprintf("/api/v1/admin/companies/%d/upload-logo", company.ID),
		nil,
		"test.png",
		"file",
		imageContent,
	)
	w := httptest.NewRecorder()

	// Execute request
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Company logo uploaded successfully", response["message"])
	assert.Contains(suite.T(), response["filename"], fmt.Sprintf("company_%d_logo", company.ID))
	assert.Contains(suite.T(), response["url"], "/uploads/logos/")

	// Verify company logo URL was updated in database
	var updatedCompany models.Company
	err = suite.testCfg.DB.First(&updatedCompany, company.ID).Error
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), updatedCompany.Logo)
	assert.Contains(suite.T(), updatedCompany.Logo, "/uploads/logos/")
}

func (suite *CompanyManagementTestSuite) TestUploadCompanyLogo_InvalidFileType() {
	// Create test company
	company := testutil.CreateTestCompany(suite.testCfg.DB)

	// Create test file with invalid extension
	fileContent := []byte("not an image")

	// Create multipart request
	req := testutil.MakeTestMultipartRequest(
		fmt.Sprintf("/api/v1/admin/companies/%d/upload-logo", company.ID),
		nil,
		"test.txt",
		"file",
		fileContent,
	)
	w := httptest.NewRecorder()

	// Execute request
	suite.router.ServeHTTP(w, req)

	// Assert response
	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "file type not allowed")
}

func (suite *CompanyManagementTestSuite) TestUploadCompanyLogo_NoFile() {
	// Create test company
	company := testutil.CreateTestCompany(suite.testCfg.DB)

	// Create request without file
	req := testutil.MakeTestRequest("POST", fmt.Sprintf("/api/v1/admin/companies/%d/upload-logo", company.ID), nil, nil)
	w := httptest.NewRecorder()

	// Execute request
	suite.router.ServeHTTP(w, req)

	// Assert response
	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "No file provided")
}

func (suite *CompanyManagementTestSuite) TestUploadCompanyLogo_CompanyNotFound() {
	// Create test file
	imageContent := []byte("fake png content")

	// Create multipart request for non-existent company
	req := testutil.MakeTestMultipartRequest(
		"/api/v1/admin/companies/999/upload-logo",
		nil,
		"test.png",
		"file",
		imageContent,
	)
	w := httptest.NewRecorder()

	// Execute request
	suite.router.ServeHTTP(w, req)

	// Assert response
	testutil.AssertErrorResponse(suite.T(), w, http.StatusNotFound, "Company not found")
}

func (suite *CompanyManagementTestSuite) TestUploadCompanyLogo_FileTooLarge() {
	// Create test company
	company := testutil.CreateTestCompany(suite.testCfg.DB)

	// Create large file content (exceed 10MB limit)
	largeContent := bytes.Repeat([]byte("x"), 11*1024*1024) // 11MB

	// Create multipart request
	req := testutil.MakeTestMultipartRequest(
		fmt.Sprintf("/api/v1/admin/companies/%d/upload-logo", company.ID),
		nil,
		"large.png",
		"file",
		largeContent,
	)
	w := httptest.NewRecorder()

	// Execute request
	suite.router.ServeHTTP(w, req)

	// Assert response
	testutil.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "file size too large")
}

// Run the test suite
func TestCompanyManagementTestSuite(t *testing.T) {
	suite.Run(t, new(CompanyManagementTestSuite))
}