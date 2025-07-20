package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kadzu/sukuk-poc-be/internal/database"
	"github.com/kadzu/sukuk-poc-be/internal/models"
)

// ListSukukSeries returns a list of all Sukuk series (READ-ONLY)
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

