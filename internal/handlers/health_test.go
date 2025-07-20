package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"sukuk-be/internal/testutil"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type HealthTestSuite struct {
	suite.Suite
	testCfg *testutil.TestConfig
	router  *gin.Engine
}

func (suite *HealthTestSuite) SetupSuite() {
	suite.testCfg = testutil.SetupTestEnvironment(suite.T())
	suite.router = gin.New()
	suite.router.GET("/health", Health)

	// Create uploads directories for testing
	os.MkdirAll("./uploads/logos", 0755)
	os.MkdirAll("./uploads/prospectus", 0755)
}

func (suite *HealthTestSuite) TearDownSuite() {
	testutil.CleanupTestEnvironment(suite.T(), suite.testCfg)
	// Clean up test uploads directory
	os.RemoveAll("./uploads")
}

func (suite *HealthTestSuite) TestHealth_Success() {
	// Create test request
	req := testutil.MakeTestRequest("GET", "/health", nil, nil)
	w := httptest.NewRecorder()

	// Execute request
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), "application/json; charset=utf-8", w.Header().Get("Content-Type"))

	// Parse response
	var response HealthStatus
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// Assert response structure
	assert.Equal(suite.T(), "healthy", response.Status)
	assert.Equal(suite.T(), "sukuk-poc-api", response.Service)
	assert.Equal(suite.T(), "1.0.0", response.Version)
	assert.NotEmpty(suite.T(), response.Timestamp)

	// Assert database health
	assert.Equal(suite.T(), "healthy", response.Database.Status)
	assert.Greater(suite.T(), response.Database.Connections, 0)
	assert.GreaterOrEqual(suite.T(), response.Database.ResponseTime.Nanoseconds(), int64(0))

	// Assert system health
	assert.Greater(suite.T(), response.System.CPUCores, 0)
	assert.Greater(suite.T(), response.System.Goroutines, 0)
	assert.Greater(suite.T(), response.System.Memory.Allocated, uint64(0))
	assert.NotEmpty(suite.T(), response.System.Uptime)

	// Assert application health
	assert.GreaterOrEqual(suite.T(), response.Application.CompaniesCount, 0)
	assert.GreaterOrEqual(suite.T(), response.Application.SukukCount, 0)

	// Assert checks
	assert.Contains(suite.T(), response.Checks, "database")
	assert.Contains(suite.T(), response.Checks, "system")
	assert.Contains(suite.T(), response.Checks, "application")
	assert.Contains(suite.T(), response.Checks, "uploads")

	assert.Equal(suite.T(), "healthy", response.Checks["database"].Status)
	assert.Equal(suite.T(), "healthy", response.Checks["system"].Status)
	assert.Equal(suite.T(), "healthy", response.Checks["application"].Status)
}

func (suite *HealthTestSuite) TestHealth_WithData() {
	// Seed test data
	testutil.SeedTestData(suite.testCfg.DB)

	// Create test request
	req := testutil.MakeTestRequest("GET", "/health", nil, nil)
	w := httptest.NewRecorder()

	// Execute request
	suite.router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Parse response
	var response HealthStatus
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// Assert that we have companies from seeded data
	assert.Equal(suite.T(), 2, response.Application.CompaniesCount)
}

func (suite *HealthTestSuite) TestHealth_UploadsDirectory() {
	// Create test request
	req := testutil.MakeTestRequest("GET", "/health", nil, nil)
	w := httptest.NewRecorder()

	// Execute request
	suite.router.ServeHTTP(w, req)

	// Parse response
	var response HealthStatus
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// The uploads directory may or may not exist in test environment
	// Just verify the check is present
	assert.Contains(suite.T(), response.Checks, "uploads")
	uploadsCheck := response.Checks["uploads"]
	assert.Contains(suite.T(), []string{"healthy", "unhealthy"}, uploadsCheck.Status)
}

func (suite *HealthTestSuite) TestHealth_ResponseTiming() {
	// Create test request
	req := testutil.MakeTestRequest("GET", "/health", nil, nil)
	w := httptest.NewRecorder()

	// Execute request
	suite.router.ServeHTTP(w, req)

	// Parse response
	var response HealthStatus
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// Assert all checks have timing information
	for checkName, check := range response.Checks {
		assert.GreaterOrEqual(suite.T(), check.Duration, int64(0), "Check %s should have valid duration", checkName)
	}
}

// Run the test suite
func TestHealthTestSuite(t *testing.T) {
	suite.Run(t, new(HealthTestSuite))
}
