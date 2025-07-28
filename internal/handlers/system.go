package handlers

import (
	"net/http"

	"sukuk-be/internal/database"
	"sukuk-be/internal/models"

	"github.com/gin-gonic/gin"
)

// System Management APIs

// GetSyncStatus returns the blockchain sync status
// @Summary Get blockchain sync status
// @Description Get the current blockchain synchronization status and last processed event
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Sync status"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security ApiKeyAuth
// @Router /admin/system/sync-status [get]
func GetSyncStatus(c *gin.Context) {
	db := database.GetDB()
	
	var systemState models.SystemState
	if err := db.Where("key = ?", "last_processed_event_id").First(&systemState).Error; err != nil {
		// If no record found, assume starting from 0
		systemState.Value = "0"
	}

	// Get total count of events in blockchain database (if accessible)
	// For now, we'll return the last processed ID
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"last_processed_event_id": systemState.Value,
			"sync_status":            "active",
			"last_updated":           systemState.UpdatedAt,
		},
	})
}

// ForceSync triggers a manual blockchain synchronization
// @Summary Force blockchain sync
// @Description Manually trigger blockchain event synchronization
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Sync triggered"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security ApiKeyAuth
// @Router /admin/system/force-sync [post]
func ForceSync(c *gin.Context) {
	// TODO: Implement actual sync triggering logic
	// This would interact with the blockchain sync service
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Blockchain sync triggered successfully",
		"status":  "sync_started",
	})
}

// GetHealthStatus returns the overall system health
// @Summary Get system health
// @Description Get overall system health including database and sync status
// @Tags System
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "System health"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /health [get]
func GetHealthStatus(c *gin.Context) {
	db := database.GetDB()
	
	// Check database connection
	sqlDB, err := db.DB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "unhealthy",
			"error":  "Database connection failed",
		})
		return
	}
	
	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": "unhealthy",
			"error":  "Database ping failed",
		})
		return
	}

	// Get basic system stats
	var sukukMetadataCount int64
	
	db.Model(&models.SukukMetadata{}).Count(&sukukMetadataCount)

	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"data": gin.H{
			"database":         "connected",
			"sukuk_metadata":   sukukMetadataCount,
		},
	})
}