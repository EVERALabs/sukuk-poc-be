package handlers

import (
	"net/http"

	"sukuk-be/internal/database"
	"sukuk-be/internal/logger"
	"sukuk-be/internal/models"
	"sukuk-be/internal/services"

	"github.com/gin-gonic/gin"
)

// GetSukukOwnedByAddress returns sukuk metadata for all sukuk owned by a specific address
// @Summary Get sukuk owned by address
// @Description Get all sukuk metadata for sukuk tokens owned by a specific wallet address
// @Tags owned-sukuk
// @Accept json
// @Produce json
// @Param address path string true "User wallet address" Example("0xf57093Ea18E5CfF6E7bB3bb770Ae9C492277A5a9")
// @Success 200 {object} OwnedSukukResponse "Owned sukuk metadata"
// @Failure 400 {object} map[string]string "Invalid address"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /owned-sukuk/{address} [get]
func GetSukukOwnedByAddress(c *gin.Context) {
	// Get address from path
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Address is required",
		})
		return
	}

	// Initialize indexer query service
	indexerService := services.NewIndexerQueryService()

	// Get all sukuk addresses owned by this user
	sukukAddresses, err := indexerService.GetSukukOwnedByAddress(address)
	if err != nil {
		logger.WithError(err).Error("Failed to fetch owned sukuk addresses")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch owned sukuk",
		})
		return
	}

	if len(sukukAddresses) == 0 {
		// User doesn't own any sukuk
		response := OwnedSukukResponse{
			Address:    address,
			TotalCount: 0,
			Sukuk:      []models.SukukMetadataListResponse{},
		}
		c.JSON(http.StatusOK, response)
		return
	}

	// Get sukuk metadata for these addresses
	var sukukMetadata []models.SukukMetadata
	result := database.GetDB().Where("contract_address IN ?", sukukAddresses).Find(&sukukMetadata)
	if result.Error != nil {
		logger.WithError(result.Error).Error("Failed to fetch sukuk metadata")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch sukuk metadata",
		})
		return
	}

	// Convert to response format with activities
	responses := make([]models.SukukMetadataListResponse, len(sukukMetadata))
	for i, sukuk := range sukukMetadata {
		response := sukuk.ToListResponse()
		
		// Get latest 10 activities for this sukuk token directly from indexer
		activities, err := indexerService.GetLatestActivities(sukuk.ContractAddress, 10)
		if err != nil {
			logger.WithError(err).Warn("Failed to fetch activities for sukuk:", sukuk.ContractAddress)
			activities = []models.ActivityEvent{} // Set empty array if error
		}
		
		response.LatestActivities = activities
		responses[i] = response
	}

	// Create response
	ownedResponse := OwnedSukukResponse{
		Address:    address,
		TotalCount: len(responses),
		Sukuk:      responses,
	}

	c.JSON(http.StatusOK, ownedResponse)
}

// OwnedSukukResponse represents the response for owned sukuk
type OwnedSukukResponse struct {
	Address    string                               `json:"address"`
	TotalCount int                                  `json:"total_count"`
	Sukuk      []models.SukukMetadataListResponse  `json:"sukuk"`
}