package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"sukuk-be/internal/database"
	"sukuk-be/internal/models"

	"github.com/gin-gonic/gin"
)

// GetYieldClaims returns a list of all yield claims with optional filtering
// @Summary List all yield claims
// @Description Get a list of all yield claims with optional filtering by status and investor address
// @Tags Yield Claims
// @Accept json
// @Produce json
// @Param status query string false "Yield claim status to filter by (pending, claimed)"
// @Param investor query string false "Investor wallet address to filter by"
// @Success 200 {object} YieldClaimListResponse "List of yield claims"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /yield-claims [get]
func GetYieldClaims(c *gin.Context) {
	var yieldClaims []models.YieldClaim

	db := database.GetDB()
	query := db.Preload("SukukSeries").Preload("SukukSeries.Company")

	// Filter by status if provided
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// Filter by investor if provided
	if investor := c.Query("investor"); investor != "" {
		query = query.Where("investor_address = ?", strings.ToLower(investor))
	}

	if err := query.Order("created_at DESC").Find(&yieldClaims).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch yield claims",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  yieldClaims,
		"count": len(yieldClaims),
	})
}

// GetYieldClaim returns details of a specific yield claim
// @Summary Get yield claim details
// @Description Get detailed information about a specific yield claim including Sukuk series, company, and investment data
// @Tags Yield Claims
// @Accept json
// @Produce json
// @Param id path int true "Yield Claim ID"
// @Success 200 {object} YieldClaimResponse "Yield claim details"
// @Failure 400 {object} ErrorResponse "Invalid ID"
// @Failure 404 {object} ErrorResponse "Yield claim not found"
// @Router /yield-claims/{id} [get]
func GetYieldClaim(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid yield claim ID",
		})
		return
	}

	var yieldClaim models.YieldClaim
	db := database.GetDB()
	if err := db.Preload("SukukSeries").Preload("SukukSeries.Company").Preload("Investment").First(&yieldClaim, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Yield claim not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": yieldClaim,
	})
}

// GetYieldClaimsByInvestor returns all yield claims for a specific investor
// @Summary Get yield claims by investor
// @Description Get all yield claims made by a specific wallet address with optional status filtering
// @Tags Yield Claims
// @Accept json
// @Produce json
// @Param address path string true "Investor wallet address"
// @Param status query string false "Yield claim status to filter by (pending, claimed)"
// @Success 200 {object} YieldClaimListResponse "Investor's yield claims"
// @Failure 400 {object} ErrorResponse "Invalid address"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /yield-claims/investor/{address} [get]
func GetYieldClaimsByInvestor(c *gin.Context) {
	address := strings.ToLower(c.Param("address"))
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid investor address",
		})
		return
	}

	var yieldClaims []models.YieldClaim
	db := database.GetDB()
	query := db.Preload("SukukSeries").Preload("SukukSeries.Company").Where("investor_address = ?", address)

	// Filter by status if provided
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&yieldClaims).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch yield claims",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  yieldClaims,
		"count": len(yieldClaims),
	})
}

// GetYieldClaimsBySukuk returns all yield claims for a specific Sukuk series
// @Summary Get yield claims by Sukuk series
// @Description Get all yield claims for a specific Sukuk series with optional status filtering
// @Tags Yield Claims
// @Accept json
// @Produce json
// @Param sukukId path int true "Sukuk Series ID"
// @Param status query string false "Yield claim status to filter by (pending, claimed)"
// @Success 200 {object} YieldClaimListResponse "Sukuk series yield claims"
// @Failure 400 {object} ErrorResponse "Invalid ID"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /yield-claims/sukuk/{sukukId} [get]
func GetYieldClaimsBySukuk(c *gin.Context) {
	sukukID, err := strconv.ParseUint(c.Param("sukukId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid sukuk series ID",
		})
		return
	}

	var yieldClaims []models.YieldClaim
	db := database.GetDB()
	query := db.Preload("SukukSeries").Where("sukuk_series_id = ?", uint(sukukID))

	// Filter by status if provided
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&yieldClaims).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch yield claims",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  yieldClaims,
		"count": len(yieldClaims),
	})
}
