package handlers

import (
	"net/http"

	"sukuk-be/internal/services"

	"github.com/gin-gonic/gin"
)

// DebugIndexerConnection tests indexer database connection and queries
// @Summary Debug indexer connection
// @Description Test connection to indexer database and query sample data
// @Tags Debug
// @Accept json
// @Produce json
// @Param address query string false "Sukuk address to test"
// @Success 200 {object} map[string]interface{} "Debug information"
// @Router /debug/indexer [get]
func DebugIndexerConnection(c *gin.Context) {
	address := c.DefaultQuery("address", "0x3de17456f9d467ad9f94e0ed66b39fA98E3E0429")
	
	// Initialize indexer query service
	indexerService := services.NewIndexerQueryService()
	
	// Test connection
	if err := indexerService.ConnectToIndexer(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to connect to indexer",
			"details": err.Error(),
		})
		return
	}
	
	// Get purchase events
	purchases, err := indexerService.GetSukukPurchases(address, 5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to query purchases",
			"details": err.Error(),
		})
		return
	}
	
	// Get redemption events
	redemptions, err := indexerService.GetRedemptionRequests(address, 5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to query redemptions", 
			"details": err.Error(),
		})
		return
	}
	
	// Get activities
	activities, err := indexerService.GetLatestActivities(address, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get activities",
			"details": err.Error(),
		})
		return
	}
	
	// Test all purchases without address filter to see what data exists
	allPurchases, err := indexerService.GetSukukPurchases("", 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get all purchases",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"address": address,
		"purchases_count": len(purchases),
		"purchases": purchases,
		"redemptions_count": len(redemptions), 
		"redemptions": redemptions,
		"activities_count": len(activities),
		"activities": activities,
		"all_purchases_count": len(allPurchases),
		"all_purchases": allPurchases,
		"connection": "successful",
	})
}