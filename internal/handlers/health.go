package handlers

import (
	"net/http"
	"os"
	"runtime"
	"time"

	"sukuk-be/internal/database"
	"sukuk-be/internal/logger"
	"sukuk-be/internal/models"

	"github.com/gin-gonic/gin"
)

type HealthStatus struct {
	Status      string                 `json:"status"`
	Timestamp   time.Time              `json:"timestamp"`
	Service     string                 `json:"service"`
	Version     string                 `json:"version"`
	Database    DatabaseHealth         `json:"database"`
	System      SystemHealth           `json:"system"`
	Application ApplicationHealth      `json:"application"`
	Checks      map[string]CheckResult `json:"checks"`
}

type DatabaseHealth struct {
	Status       string        `json:"status"`
	ResponseTime time.Duration `json:"response_time_ms"`
	Connections  int           `json:"active_connections"`
	Error        string        `json:"error,omitempty"`
}

type SystemHealth struct {
	Memory     MemoryStats `json:"memory"`
	Goroutines int         `json:"goroutines"`
	CPUCores   int         `json:"cpu_cores"`
	Uptime     string      `json:"uptime"`
}

type MemoryStats struct {
	Allocated    uint64 `json:"allocated_mb"`
	TotalAlloc   uint64 `json:"total_alloc_mb"`
	SystemMemory uint64 `json:"system_mb"`
	GCRuns       uint32 `json:"gc_runs"`
}

type ApplicationHealth struct {
	CompaniesCount int  `json:"companies_count"`
	SukukCount     int  `json:"sukuk_series_count"`
	UploadsDir     bool `json:"uploads_directory_writable"`
}

type CheckResult struct {
	Status   string `json:"status"`
	Message  string `json:"message,omitempty"`
	Duration int64  `json:"duration_ms"`
}

var startTime = time.Now()

// Health godoc
// @Summary Health Check
// @Description Get the health status of the API including database, system, and application metrics
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthStatus "Service is healthy"
// @Success 503 {object} HealthStatus "Service is unhealthy"
// @Router /health [get]
func Health(c *gin.Context) {
	start := time.Now()

	// Initialize health status
	health := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now().UTC(),
		Service:   "sukuk-poc-api",
		Version:   "1.0.0",
		Checks:    make(map[string]CheckResult),
	}

	// Check database health
	dbStart := time.Now()
	dbHealth := checkDatabaseHealth()
	health.Database = dbHealth
	health.Checks["database"] = CheckResult{
		Status:   dbHealth.Status,
		Message:  dbHealth.Error,
		Duration: time.Since(dbStart).Milliseconds(),
	}

	// Check system health
	health.System = getSystemHealth()
	health.Checks["system"] = CheckResult{
		Status:   "healthy",
		Duration: 0, // System checks are instant
	}

	// Check application health
	appStart := time.Now()
	appHealth := checkApplicationHealth()
	health.Application = appHealth
	health.Checks["application"] = CheckResult{
		Status:   "healthy",
		Duration: time.Since(appStart).Milliseconds(),
	}

	// Check uploads directory
	uploadsStart := time.Now()
	uploadsHealthy := checkUploadsDirectory()
	health.Checks["uploads"] = CheckResult{
		Status:   map[bool]string{true: "healthy", false: "unhealthy"}[uploadsHealthy],
		Duration: time.Since(uploadsStart).Milliseconds(),
	}

	// Determine overall status
	overallStatus := "healthy"
	httpStatus := http.StatusOK

	if health.Database.Status != "healthy" {
		overallStatus = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	} else if !uploadsHealthy {
		overallStatus = "degraded"
		httpStatus = http.StatusOK // Still serve traffic but with warnings
	}

	health.Status = overallStatus

	// Log health check
	logger.WithFields(map[string]interface{}{
		"status":           overallStatus,
		"database_status":  health.Database.Status,
		"response_time_ms": time.Since(start).Milliseconds(),
	}).Info("Health check performed")

	c.JSON(httpStatus, health)
}

func checkDatabaseHealth() DatabaseHealth {
	dbHealth := DatabaseHealth{
		Status: "healthy",
	}

	start := time.Now()

	// Check basic connectivity
	if err := database.Health(); err != nil {
		dbHealth.Status = "unhealthy"
		dbHealth.Error = err.Error()
		dbHealth.ResponseTime = time.Since(start)
		return dbHealth
	}

	dbHealth.ResponseTime = time.Since(start)

	// Get database stats
	db := database.GetDB()
	if sqlDB, err := db.DB(); err == nil {
		stats := sqlDB.Stats()
		dbHealth.Connections = stats.OpenConnections
	}

	return dbHealth
}

func getSystemHealth() SystemHealth {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemHealth{
		Memory: MemoryStats{
			Allocated:    m.Alloc / 1024 / 1024,      // Convert to MB
			TotalAlloc:   m.TotalAlloc / 1024 / 1024, // Convert to MB
			SystemMemory: m.Sys / 1024 / 1024,        // Convert to MB
			GCRuns:       m.NumGC,
		},
		Goroutines: runtime.NumGoroutine(),
		CPUCores:   runtime.NumCPU(),
		Uptime:     time.Since(startTime).String(),
	}
}

func checkApplicationHealth() ApplicationHealth {
	appHealth := ApplicationHealth{}

	db := database.GetDB()

	// Count companies
	var companyCount int64
	db.Model(&models.Company{}).Count(&companyCount)
	appHealth.CompaniesCount = int(companyCount)

	// Count sukuk series
	var sukukCount int64
	db.Model(&models.SukukSeries{}).Count(&sukukCount)
	appHealth.SukukCount = int(sukukCount)

	// Check uploads directory
	appHealth.UploadsDir = checkUploadsDirectory()

	return appHealth
}

func checkUploadsDirectory() bool {
	// Check if uploads directories exist and are writable
	dirs := []string{"./uploads/logos", "./uploads/prospectus"}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return false
		}

		// Try to create a temporary file to test write permissions
		testFile := dir + "/.health_check"
		if file, err := os.Create(testFile); err != nil {
			return false
		} else {
			file.Close()
			os.Remove(testFile)
		}
	}

	return true
}
