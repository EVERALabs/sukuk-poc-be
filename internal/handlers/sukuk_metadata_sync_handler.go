package handlers

import (
	"net/http"
	"strconv"

	"sukuk-be/internal/logger"
	"sukuk-be/internal/services"

	"github.com/gin-gonic/gin"
)

// TriggerSukukMetadataSync manually triggers sync for a specific sukuk
// @Summary Trigger sukuk metadata sync
// @Description Manually sync metadata for a specific sukuk from indexer
// @Tags sukuk-metadata
// @Accept json
// @Produce json
// @Param tokenId query int true "Token ID"
// @Param contractAddress query string true "Contract Address"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /sukuk-metadata/sync [post]
func TriggerSukukMetadataSync(c *gin.Context) {
	// Get parameters
	tokenIDStr := c.Query("tokenId")
	contractAddress := c.Query("contractAddress")
	
	// Validate token ID
	if tokenIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "tokenId is required",
		})
		return
	}
	
	tokenID, err := strconv.ParseInt(tokenIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid tokenId format",
		})
		return
	}
	
	// Validate contract address
	if contractAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "contractAddress is required",
		})
		return
	}
	
	// Create sync service instance
	syncService := services.NewSukukMetadataSyncService(0) // 0 interval for one-time sync
	
	// Sync specific sukuk
	if err := syncService.SyncSpecificSukuk(tokenID, contractAddress); err != nil {
		logger.WithError(err).Error("Failed to sync sukuk metadata")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to sync sukuk metadata",
			"details": err.Error(),
		})
		return
	}
	
	logger.WithFields(map[string]interface{}{
		"token_id": tokenID,
		"contract_address": contractAddress,
	}).Info("Sukuk metadata sync triggered successfully")
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Sukuk metadata sync completed successfully",
		"token_id": tokenID,
		"contract_address": contractAddress,
	})
}

// ListSukukCreationTables lists all available sukuk creation tables in the indexer
// @Summary List sukuk creation tables
// @Description Get all available sukuk creation tables from the indexer
// @Tags sukuk-metadata
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /sukuk-metadata/tables [get]
func ListSukukCreationTables(c *gin.Context) {
	// Create sync service instance
	syncService := services.NewSukukMetadataSyncService(0)
	
	// Get all sukuk creation tables
	tables, err := syncService.FindAllSukukCreationTables()
	if err != nil {
		logger.WithError(err).Error("Failed to get sukuk creation tables")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get sukuk creation tables",
			"details": err.Error(),
		})
		return
	}
	
	// Get latest table
	latestTable, err := syncService.FindLatestSukukCreationTable()
	if err != nil {
		logger.WithError(err).Error("Failed to get latest sukuk creation table")
		latestTable = "error getting latest"
	}
	
	c.JSON(http.StatusOK, gin.H{
		"tables": tables,
		"latest_table": latestTable,
		"total_count": len(tables),
	})
}