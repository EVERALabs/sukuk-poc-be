package handlers

import (
	"net/http"
	"strconv"

	"sukuk-be/internal/logger"
	"sukuk-be/internal/models"
	"sukuk-be/internal/services"

	"github.com/gin-gonic/gin"
)

// GetRiwayatByAddress returns transaction history for a specific address
// @Summary Get transaction history by address
// @Description Get all blockchain activities (purchases and redemptions) for a specific user address
// @Tags transaction-history
// @Accept json
// @Produce json
// @Param address path string true "User wallet address" Example("0xf57093Ea18E5CfF6E7bB3bb770Ae9C492277A5a9")
// @Param limit query integer false "Number of activities to retrieve" default(50) Example(20)
// @Success 200 {object} RiwayatResponse "Transaction history"
// @Failure 400 {object} map[string]string "Invalid address or limit"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /transaction-history/{address} [get]
func GetRiwayatByAddress(c *gin.Context) {
	// Get address from path
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Address is required",
		})
		return
	}

	// Get limit from query parameter
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid limit parameter",
		})
		return
	}

	// Initialize indexer query service
	indexerService := services.NewIndexerQueryService()

	// Get all activities for this address
	activities, err := indexerService.GetActivitiesByAddress(address, limit)
	if err != nil {
		logger.WithError(err).Error("Failed to fetch user activities")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch transaction history",
		})
		return
	}

	// Create response
	response := RiwayatResponse{
		Address:        address,
		TotalCount:     len(activities),
		Activities:     activities,
	}

	c.JSON(http.StatusOK, response)
}

// RiwayatResponse represents the response for transaction history
type RiwayatResponse struct {
	Address        string                   `json:"address"`
	TotalCount     int                      `json:"total_count"`
	Activities     []models.ActivityEvent   `json:"activities"`
}