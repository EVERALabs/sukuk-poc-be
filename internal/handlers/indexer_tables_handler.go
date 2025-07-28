package handlers

import (
	"net/http"

	"sukuk-be/internal/logger"
	"sukuk-be/internal/models"
	"sukuk-be/internal/services"

	"github.com/gin-gonic/gin"
)

// ListIndexerTables returns all discovered indexer tables with metadata
// @Summary List indexer tables
// @Description Get all discovered hash-prefixed indexer tables with metadata and row counts
// @Tags debug
// @Accept json
// @Produce json
// @Success 200 {object} models.IndexerTablesResponse "Discovered indexer tables"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /debug/indexer-tables [get]
func ListIndexerTables(c *gin.Context) {
	// Initialize table discovery service
	tableService := services.NewIndexerTableService()
	
	if err := tableService.ConnectToIndexer(); err != nil {
		logger.WithError(err).Error("Failed to connect to indexer")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to connect to indexer",
		})
		return
	}

	// Discover all tables
	discoveredTables, err := tableService.DiscoverAllTables()
	if err != nil {
		logger.WithError(err).Error("Failed to discover indexer tables")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to discover indexer tables",
		})
		return
	}

	// Get latest tables mapping
	latestTables, err := tableService.GetAllLatestTables()
	if err != nil {
		logger.WithError(err).Error("Failed to get latest tables")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get latest tables",
		})
		return
	}

	// Get available event types
	eventTypes, err := tableService.GetAvailableEventTypes()
	if err != nil {
		logger.WithError(err).Error("Failed to get available event types")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get available event types",
		})
		return
	}

	// Convert to API response format
	response := models.IndexerTablesResponse{
		TotalTables:     len(discoveredTables),
		AvailableEvents: eventTypes,
		Tables:          make([]models.IndexerTableInfo, len(discoveredTables)),
		LatestTables:    latestTables,
	}

	// Populate table information with row counts
	for i, table := range discoveredTables {
		tableInfo := models.IndexerTableInfo{
			EventType:  table.EventType,
			TableName:  table.FullName,
			HashPrefix: table.HashPrefix,
		}

		// Get row count for each table
		rowCount, err := tableService.GetTableRowCount(table.FullName)
		if err == nil {
			tableInfo.RowCount = rowCount
		}

		response.Tables[i] = tableInfo
	}

	c.JSON(http.StatusOK, response)
}

// ValidateIndexerTables validates the structure of all discovered tables
// @Summary Validate indexer tables
// @Description Validate that all discovered indexer tables have expected columns
// @Tags debug
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Validation results"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /debug/indexer-tables/validate [get]
func ValidateIndexerTables(c *gin.Context) {
	// Initialize table discovery service
	tableService := services.NewIndexerTableService()
	
	if err := tableService.ConnectToIndexer(); err != nil {
		logger.WithError(err).Error("Failed to connect to indexer")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to connect to indexer",
		})
		return
	}

	// Get latest tables for validation
	latestTables, err := tableService.GetAllLatestTables()
	if err != nil {
		logger.WithError(err).Error("Failed to get latest tables")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get latest tables",
		})
		return
	}

	validationResults := make(map[string]interface{})
	validTables := 0
	invalidTables := 0

	// Validate each latest table
	for eventType, tableName := range latestTables {
		err := tableService.ValidateTableStructure(tableName, eventType)
		if err != nil {
			validationResults[eventType] = gin.H{
				"table":  tableName,
				"valid":  false,
				"error":  err.Error(),
			}
			invalidTables++
		} else {
			validationResults[eventType] = gin.H{
				"table": tableName,
				"valid": true,
			}
			validTables++
		}
	}

	response := gin.H{
		"total_tables":    len(latestTables),
		"valid_tables":    validTables,
		"invalid_tables":  invalidTables,
		"validation_results": validationResults,
		"all_valid":       invalidTables == 0,
	}

	c.JSON(http.StatusOK, response)
}

// GetTableDetails returns detailed information about a specific table
// @Summary Get table details
// @Description Get detailed information about a specific indexer table including column structure
// @Tags debug
// @Accept json
// @Produce json
// @Param table_name path string true "Table name" Example("f243__sukuk_purchase")
// @Success 200 {object} map[string]interface{} "Table details"
// @Failure 400 {object} map[string]string "Invalid table name"
// @Failure 404 {object} map[string]string "Table not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /debug/indexer-tables/{table_name} [get]
func GetTableDetails(c *gin.Context) {
	tableName := c.Param("table_name")
	if tableName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Table name is required",
		})
		return
	}

	// Initialize table discovery service
	tableService := services.NewIndexerTableService()
	
	if err := tableService.ConnectToIndexer(); err != nil {
		logger.WithError(err).Error("Failed to connect to indexer")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to connect to indexer",
		})
		return
	}

	// Check if table exists
	exists, err := tableService.CheckTableExists(tableName)
	if err != nil {
		logger.WithError(err).Error("Failed to check table existence")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check table existence",
		})
		return
	}

	if !exists {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Table not found",
		})
		return
	}

	// Get row count
	rowCount, err := tableService.GetTableRowCount(tableName)
	if err != nil {
		logger.WithError(err).Error("Failed to get row count")
		rowCount = -1 // Indicate error in getting count
	}

	// TODO: Add column information query
	// For now, return basic information
	response := gin.H{
		"table_name": tableName,
		"exists":     true,
		"row_count":  rowCount,
		"schema":     "public",
	}

	c.JSON(http.StatusOK, response)
}

// GetHashPrefixTables returns all tables with a specific hash prefix
// @Summary Get tables by hash prefix
// @Description Get all tables that share the same hash prefix (deployment)
// @Tags debug
// @Accept json
// @Produce json
// @Param hash_prefix path string true "Hash prefix" Example("f243")
// @Success 200 {object} map[string]interface{} "Tables with hash prefix"
// @Failure 400 {object} map[string]string "Invalid hash prefix"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /debug/indexer-tables/prefix/{hash_prefix} [get]
func GetHashPrefixTables(c *gin.Context) {
	hashPrefix := c.Param("hash_prefix")
	if hashPrefix == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Hash prefix is required",
		})
		return
	}

	// Initialize table discovery service
	tableService := services.NewIndexerTableService()
	
	if err := tableService.ConnectToIndexer(); err != nil {
		logger.WithError(err).Error("Failed to connect to indexer")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to connect to indexer",
		})
		return
	}

	// Get tables with this hash prefix
	tables, err := tableService.GetTablesByHashPrefix(hashPrefix)
	if err != nil {
		logger.WithError(err).Error("Failed to get tables by hash prefix")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get tables by hash prefix",
		})
		return
	}

	// Convert to API response format
	apiTables := make([]models.IndexerTableInfo, len(tables))
	for i, table := range tables {
		tableInfo := models.IndexerTableInfo{
			EventType:  table.EventType,
			TableName:  table.FullName,
			HashPrefix: table.HashPrefix,
		}

		// Get row count
		rowCount, err := tableService.GetTableRowCount(table.FullName)
		if err == nil {
			tableInfo.RowCount = rowCount
		}

		apiTables[i] = tableInfo
	}

	response := gin.H{
		"hash_prefix":  hashPrefix,
		"total_tables": len(apiTables),
		"tables":       apiTables,
	}

	c.JSON(http.StatusOK, response)
}