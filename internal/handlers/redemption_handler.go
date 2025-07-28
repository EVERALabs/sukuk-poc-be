package handlers

import (
	"net/http"
	"strconv"

	"sukuk-be/internal/logger"
	"sukuk-be/internal/models"
	"sukuk-be/internal/services"

	"github.com/gin-gonic/gin"
)

// GetAllRedemptions returns all redemption requests with their approval status
// @Summary Get all redemptions
// @Description Get all redemption requests with their approval status, supports pagination
// @Tags redemptions
// @Accept json
// @Produce json
// @Param limit query int false "Number of redemptions to return" default(50) minimum(1) maximum(200)
// @Param offset query int false "Number of redemptions to skip" default(0) minimum(0)
// @Param status query string false "Filter by status" Enums(requested, approved, rejected, completed)
// @Success 200 {object} models.RedemptionListResponse "List of redemptions with status"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /redemptions [get]
func GetAllRedemptions(c *gin.Context) {
	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}

	offsetStr := c.DefaultQuery("offset", "0")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Initialize redemption service
	redemptionService := services.NewRedemptionService()

	// Get all redemptions
	redemptions, err := redemptionService.GetAllRedemptions(limit, offset)
	if err != nil {
		logger.WithError(err).Error("Failed to get all redemptions")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get redemptions",
		})
		return
	}

	// Filter by status if specified
	status := c.Query("status")
	if status != "" {
		filteredRedemptions := []models.RedemptionRequest{}
		for _, r := range redemptions.Redemptions {
			if string(r.Status) == status {
				filteredRedemptions = append(filteredRedemptions, r)
			}
		}
		redemptions.Redemptions = filteredRedemptions
		redemptions.TotalCount = len(filteredRedemptions)
	}

	c.JSON(http.StatusOK, redemptions)
}

// GetRedemptionsByUser returns redemptions for a specific user
// @Summary Get user redemptions
// @Description Get all redemption requests and approvals for a specific user
// @Tags redemptions
// @Accept json
// @Produce json
// @Param address path string true "User wallet address" Example("0xf57093Ea18E5CfF6E7bB3bb770Ae9C492277A5a9")
// @Success 200 {object} models.RedemptionListResponse "User's redemptions"
// @Failure 400 {object} map[string]string "Invalid address"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /redemptions/user/{address} [get]
func GetRedemptionsByUser(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Address is required",
		})
		return
	}

	// Initialize redemption service
	redemptionService := services.NewRedemptionService()

	// Get user's redemptions
	redemptions, err := redemptionService.GetRedemptionsByUser(address)
	if err != nil {
		logger.WithError(err).Error("Failed to get user redemptions")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user redemptions",
		})
		return
	}

	c.JSON(http.StatusOK, redemptions)
}

// GetRedemptionsBySukuk returns redemptions for a specific sukuk
// @Summary Get sukuk redemptions
// @Description Get all redemption requests and approvals for a specific sukuk
// @Tags redemptions
// @Accept json
// @Produce json
// @Param sukuk_address path string true "Sukuk contract address" Example("0x02ba44871BD555d6ebD541e2820796F9b88cBF75")
// @Success 200 {object} models.RedemptionListResponse "Sukuk's redemptions"
// @Failure 400 {object} map[string]string "Invalid sukuk address"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /redemptions/sukuk/{sukuk_address} [get]
func GetRedemptionsBySukuk(c *gin.Context) {
	sukukAddress := c.Param("sukuk_address")
	if sukukAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Sukuk address is required",
		})
		return
	}

	// Initialize redemption service
	redemptionService := services.NewRedemptionService()

	// Get sukuk's redemptions
	redemptions, err := redemptionService.GetRedemptionsBySukuk(sukukAddress)
	if err != nil {
		logger.WithError(err).Error("Failed to get sukuk redemptions")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get sukuk redemptions",
		})
		return
	}

	c.JSON(http.StatusOK, redemptions)
}

// GetRedemptionStats returns overall redemption statistics
// @Summary Get redemption statistics
// @Description Get comprehensive statistics about all redemptions
// @Tags redemptions
// @Accept json
// @Produce json
// @Success 200 {object} models.RedemptionStatsResponse "Redemption statistics"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /redemptions/stats [get]
func GetRedemptionStats(c *gin.Context) {
	// Initialize redemption service
	redemptionService := services.NewRedemptionService()

	// Get redemption statistics
	stats, err := redemptionService.GetRedemptionStats()
	if err != nil {
		logger.WithError(err).Error("Failed to get redemption stats")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get redemption statistics",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}


// GetRedemptionByID returns a specific redemption by request ID
// @Summary Get redemption by ID
// @Description Get detailed information about a specific redemption request
// @Tags redemptions
// @Accept json
// @Produce json
// @Param request_id path string true "Redemption request ID"
// @Success 200 {object} models.RedemptionRequest "Redemption details"
// @Failure 400 {object} map[string]string "Invalid request ID"
// @Failure 404 {object} map[string]string "Redemption not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /redemptions/{request_id} [get]
func GetRedemptionByID(c *gin.Context) {
	requestID := c.Param("request_id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Request ID is required",
		})
		return
	}

	// Initialize redemption service
	redemptionService := services.NewRedemptionService()

	// Get all redemptions and find the specific one
	// Note: This is not the most efficient, but works for MVP
	// In production, you'd want a direct lookup method
	allRedemptions, err := redemptionService.GetAllRedemptions(1000, 0)
	if err != nil {
		logger.WithError(err).Error("Failed to get redemptions")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get redemption",
		})
		return
	}

	// Find the specific redemption
	for _, r := range allRedemptions.Redemptions {
		if r.RequestID == requestID {
			c.JSON(http.StatusOK, r)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"error": "Redemption request not found",
	})
}