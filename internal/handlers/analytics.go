package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kadzu/sukuk-poc-be/internal/database"
	"github.com/kadzu/sukuk-poc-be/internal/models"
)

// GetPlatformStats returns platform-wide statistics (READ-ONLY)
func GetPlatformStats(c *gin.Context) {
	db := database.GetDB()
	
	// Get total companies
	var totalCompanies int64
	db.Model(&models.Company{}).Where("is_active = ?", true).Count(&totalCompanies)
	
	// Get total active sukuk series
	var totalSukukSeries int64
	db.Model(&models.SukukSeries{}).Where("status = ?", "active").Count(&totalSukukSeries)
	
	// Get total active investments
	var totalInvestors int64
	db.Model(&models.Investment{}).Where("status = ?", "active").Count(&totalInvestors)
	
	// Get unique investor count
	var uniqueInvestors int64
	db.Model(&models.Investment{}).Distinct("investor_address").Where("status = ?", "active").Count(&uniqueInvestors)
	
	// Get pending yield claims
	var pendingYields int64
	db.Model(&models.YieldClaim{}).Where("status = ?", "pending").Count(&pendingYields)
	
	// Get pending redemptions
	var pendingRedemptions int64
	db.Model(&models.Redemption{}).Where("status = ?", "requested").Count(&pendingRedemptions)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"total_companies": totalCompanies,
			"total_sukuk_series": totalSukukSeries,
			"total_investments": totalInvestors,
			"unique_investors": uniqueInvestors,
			"pending_yields": pendingYields,
			"pending_redemptions": pendingRedemptions,
		},
	})
}

// GetVaultBalance returns IDRX vault balance for a specific series (READ-ONLY)
func GetVaultBalance(c *gin.Context) {
	seriesID, err := strconv.ParseUint(c.Param("seriesId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid series ID",
		})
		return
	}

	// Check if series exists
	var sukukSeries models.SukukSeries
	db := database.GetDB()
	if err := db.First(&sukukSeries, uint(seriesID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Sukuk series not found",
		})
		return
	}

	// Calculate vault balance (this would normally come from smart contract)
	// For now, return outstanding supply as vault balance
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"series_id": seriesID,
			"series_name": sukukSeries.Name,
			"vault_balance": sukukSeries.OutstandingSupply,
			"total_supply": sukukSeries.TotalSupply,
			"utilization_rate": "0.0", // Calculate based on vault_balance/total_supply
		},
	})
}