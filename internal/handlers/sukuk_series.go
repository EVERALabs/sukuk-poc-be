package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kadzu/sukuk-poc-be/internal/database"
	"github.com/kadzu/sukuk-poc-be/internal/models"
)

// ListSukukSeries returns a list of all Sukuk series (READ-ONLY)
// @Summary List all Sukuk series
// @Description Get a list of all Sukuk series with optional filtering by company and status
// @Tags Sukuk Series
// @Accept json
// @Produce json
// @Param company_id query string false "Company ID to filter by"
// @Param status query string false "Status to filter by (active, paused, matured)"
// @Success 200 {object} SukukSeriesListResponse "List of Sukuk series"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /sukuks [get]
func ListSukukSeries(c *gin.Context) {
	var sukukSeries []models.SukukSeries
	
	db := database.GetDB()
	query := db.Preload("Company")

	// Filter by company if provided
	if companyID := c.Query("company_id"); companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}

	// Filter by status if provided
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&sukukSeries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch sukuk series",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": sukukSeries,
		"count": len(sukukSeries),
		"meta": gin.H{
			"total": len(sukukSeries),
		},
	})
}

// GetSukukSeries returns details of a specific Sukuk series (READ-ONLY)
// @Summary Get Sukuk series details
// @Description Get detailed information about a specific Sukuk series including company, investments, yield claims, and redemptions
// @Tags Sukuk Series
// @Accept json
// @Produce json
// @Param id path int true "Sukuk Series ID"
// @Success 200 {object} SukukSeriesResponse "Sukuk series details"
// @Failure 400 {object} ErrorResponse "Invalid ID"
// @Failure 404 {object} ErrorResponse "Sukuk series not found"
// @Router /sukuks/{id} [get]
func GetSukukSeries(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid sukuk series ID",
		})
		return
	}

	var sukukSeries models.SukukSeries
	db := database.GetDB()
	if err := db.Preload("Company").Preload("Investments").Preload("YieldClaims").Preload("Redemptions").First(&sukukSeries, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Sukuk series not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": sukukSeries,
	})
}

// GetSukukMetrics returns performance metrics for a specific Sukuk series
// @Summary Get Sukuk series metrics
// @Description Get performance metrics including total investors, investment amount, pending yields, and redemptions
// @Tags Sukuk Series
// @Accept json
// @Produce json
// @Param id path int true "Sukuk Series ID"
// @Success 200 {object} SukukMetricsResponse "Sukuk series metrics"
// @Failure 400 {object} ErrorResponse "Invalid ID"
// @Router /sukuks/{id}/metrics [get]
func GetSukukMetrics(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid sukuk series ID",
		})
		return
	}

	db := database.GetDB()
	
	// Get total investments
	var totalInvestors int64
	var totalInvestment string = "0"
	
	db.Model(&models.Investment{}).Where("sukuk_series_id = ? AND status = ?", uint(id), "active").Count(&totalInvestors)
	
	// Get pending yields
	var pendingYields int64
	db.Model(&models.YieldClaim{}).Where("sukuk_series_id = ? AND status = ?", uint(id), "pending").Count(&pendingYields)
	
	// Get pending redemptions
	var pendingRedemptions int64
	db.Model(&models.Redemption{}).Where("sukuk_series_id = ? AND status = ?", uint(id), "requested").Count(&pendingRedemptions)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"total_investors": totalInvestors,
			"total_investment": totalInvestment,
			"pending_yields": pendingYields,
			"pending_redemptions": pendingRedemptions,
		},
	})
}

// GetSukukHolders returns current holders of a specific Sukuk
// @Summary Get Sukuk holders
// @Description Get a list of current active investors holding the specified Sukuk series
// @Tags Sukuk Series
// @Accept json
// @Produce json
// @Param id path int true "Sukuk Series ID"
// @Success 200 {object} SukukHoldersResponse "List of active investments"
// @Failure 400 {object} ErrorResponse "Invalid ID"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /sukuks/{id}/holders [get]
func GetSukukHolders(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid sukuk series ID",
		})
		return
	}

	var investments []models.Investment
	db := database.GetDB()
	if err := db.Preload("SukukSeries").Where("sukuk_series_id = ? AND status = ?", uint(id), "active").Find(&investments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch sukuk holders",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": investments,
		"count": len(investments),
	})
}

