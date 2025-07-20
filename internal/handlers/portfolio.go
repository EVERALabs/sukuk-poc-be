package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kadzu/sukuk-poc-be/internal/database"
	"github.com/kadzu/sukuk-poc-be/internal/models"
)

// GetPortfolio returns user's complete portfolio (READ-ONLY)
func GetPortfolio(c *gin.Context) {
	address := strings.ToLower(c.Param("address"))
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid wallet address",
		})
		return
	}

	db := database.GetDB()
	
	// Get all active investments
	var investments []models.Investment
	if err := db.Preload("SukukSeries").Preload("SukukSeries.Company").Where("investor_address = ? AND status = ?", address, "active").Find(&investments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch investments",
		})
		return
	}

	// Get pending yield claims
	var pendingYields []models.YieldClaim
	if err := db.Preload("SukukSeries").Where("investor_address = ? AND status = ?", address, "pending").Find(&pendingYields).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch pending yields",
		})
		return
	}

	// Get redemption requests
	var redemptions []models.Redemption
	if err := db.Preload("SukukSeries").Where("investor_address = ?", address).Find(&redemptions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch redemptions",
		})
		return
	}

	// Calculate summary
	totalInvestments := len(investments)
	totalPendingYields := len(pendingYields)
	totalRedemptions := len(redemptions)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"address": address,
			"summary": gin.H{
				"total_investments": totalInvestments,
				"pending_yields": totalPendingYields,
				"total_redemptions": totalRedemptions,
			},
			"investments": investments,
			"pending_yields": pendingYields,
			"redemptions": redemptions,
		},
	})
}

// GetInvestmentHistory returns user's investment history (READ-ONLY)
func GetInvestmentHistory(c *gin.Context) {
	address := strings.ToLower(c.Param("address"))
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid wallet address",
		})
		return
	}

	var investments []models.Investment
	db := database.GetDB()
	query := db.Preload("SukukSeries").Preload("SukukSeries.Company").Where("investor_address = ?", address)

	// Filter by status if provided
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&investments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch investment history",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": investments,
		"count": len(investments),
	})
}

// GetYieldHistory returns user's yield claim history (READ-ONLY)
func GetYieldHistory(c *gin.Context) {
	address := strings.ToLower(c.Param("address"))
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid wallet address",
		})
		return
	}

	var yieldClaims []models.YieldClaim
	db := database.GetDB()
	query := db.Preload("SukukSeries").Where("investor_address = ?", address)

	// Filter by status if provided
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&yieldClaims).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch yield history",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": yieldClaims,
		"count": len(yieldClaims),
	})
}

// GetPendingYields returns user's pending yield claims (READ-ONLY)
func GetPendingYields(c *gin.Context) {
	address := strings.ToLower(c.Param("address"))
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid wallet address",
		})
		return
	}

	var pendingYields []models.YieldClaim
	db := database.GetDB()
	if err := db.Preload("SukukSeries").Preload("SukukSeries.Company").Where("investor_address = ? AND status = ?", address, "pending").Find(&pendingYields).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch pending yields",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": pendingYields,
		"count": len(pendingYields),
	})
}

// GetRedemptionHistory returns user's redemption history (READ-ONLY)
func GetRedemptionHistory(c *gin.Context) {
	address := strings.ToLower(c.Param("address"))
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid wallet address",
		})
		return
	}

	var redemptions []models.Redemption
	db := database.GetDB()
	query := db.Preload("SukukSeries").Preload("SukukSeries.Company").Where("investor_address = ?", address)

	// Filter by status if provided
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&redemptions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch redemption history",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": redemptions,
		"count": len(redemptions),
	})
}