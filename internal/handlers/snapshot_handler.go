package handlers

import (
	"net/http"
	"strconv"

	"sukuk-be/internal/logger"
	"sukuk-be/internal/models"
	"sukuk-be/internal/services"

	"github.com/gin-gonic/gin"
)

// GetSukukSnapshots returns snapshot events for a specific sukuk
// @Summary Get sukuk snapshots
// @Description Get balance snapshots for a sukuk token used for yield distribution calculations
// @Tags snapshots
// @Accept json
// @Produce json
// @Param sukukAddress path string true "Sukuk contract address" Example("0x71D7C963E607eeDAfAA7Ef8f8c92bBb878090650")
// @Param limit query int false "Number of snapshots to return (default: 10, max: 100)"
// @Param snapshot_id query int false "Filter by specific snapshot ID"
// @Success 200 {object} SnapshotsResponse "Sukuk snapshots"
// @Failure 400 {object} map[string]string "Invalid parameters"
// @Failure 404 {object} map[string]string "Sukuk not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /sukuk/{sukukAddress}/snapshots [get]
func GetSukukSnapshots(c *gin.Context) {
	// Get sukuk address from path
	sukukAddress := c.Param("sukukAddress")
	if sukukAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Sukuk address is required",
		})
		return
	}

	// Check for specific snapshot ID
	snapshotIdStr := c.Query("snapshot_id")
	if snapshotIdStr != "" {
		snapshotId, err := strconv.ParseInt(snapshotIdStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid snapshot_id format",
			})
			return
		}

		// Get specific snapshot
		indexerService := services.NewIndexerQueryService()
		snapshot, err := indexerService.GetSnapshotById(sukukAddress, snapshotId)
		if err != nil {
			logger.WithError(err).Error("Failed to fetch snapshot by ID")
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Snapshot not found",
			})
			return
		}

		c.JSON(http.StatusOK, snapshot)
		return
	}

	// Get limit parameter
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			if parsedLimit > 0 && parsedLimit <= 100 {
				limit = parsedLimit
			}
		}
	}

	// Initialize indexer query service
	indexerService := services.NewIndexerQueryService()

	// Get snapshots for this sukuk
	snapshots, err := indexerService.GetSnapshots(sukukAddress, limit)
	if err != nil {
		logger.WithError(err).Error("Failed to fetch sukuk snapshots")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch snapshots",
		})
		return
	}

	// Create response
	response := SnapshotsResponse{
		SukukAddress: sukukAddress,
		TotalCount:   len(snapshots),
		Snapshots:    snapshots,
	}

	c.JSON(http.StatusOK, response)
}

// GetAllSnapshots returns snapshot events for all sukuk
// @Summary Get all snapshots
// @Description Get balance snapshots for all sukuk tokens
// @Tags snapshots
// @Accept json
// @Produce json
// @Param limit query int false "Number of snapshots to return (default: 50, max: 200)"
// @Success 200 {object} AllSnapshotsResponse "All snapshots"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /snapshots [get]
func GetAllSnapshots(c *gin.Context) {
	// Get limit parameter
	limit := 50
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			if parsedLimit > 0 && parsedLimit <= 200 {
				limit = parsedLimit
			}
		}
	}

	// Initialize indexer query service
	indexerService := services.NewIndexerQueryService()

	// Get all snapshots (pass empty string for all sukuk)
	snapshots, err := indexerService.GetAllSnapshots(limit)
	if err != nil {
		logger.WithError(err).Error("Failed to fetch all snapshots")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch snapshots",
		})
		return
	}

	// Create response
	response := AllSnapshotsResponse{
		TotalCount: len(snapshots),
		Snapshots:  snapshots,
	}

	c.JSON(http.StatusOK, response)
}

// SnapshotsResponse represents the response for sukuk snapshots
type SnapshotsResponse struct {
	SukukAddress string                  `json:"sukuk_address"`
	TotalCount   int                     `json:"total_count"`
	Snapshots    []models.SnapshotEvent  `json:"snapshots"`
}

// AllSnapshotsResponse represents the response for all snapshots
type AllSnapshotsResponse struct {
	TotalCount int                     `json:"total_count"`
	Snapshots  []models.SnapshotEvent  `json:"snapshots"`
}