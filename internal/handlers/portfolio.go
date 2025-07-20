package handlers

import (
	"net/http"
	"strings"

	"sukuk-be/internal/database"
	"sukuk-be/internal/models"

	"github.com/gin-gonic/gin"
)

// GetPortfolio returns user's complete portfolio (READ-ONLY)
// @Summary Get user portfolio
// @Description Get complete portfolio overview for a wallet address including investments, pending yields, and redemptions
// @Tags Portfolio
// @Accept json
// @Produce json
// @Param address path string true "Wallet address"
// @Success 200 {object} PortfolioResponse "Portfolio data"
// @Failure 400 {object} ErrorResponse "Invalid address"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /portfolio/{address} [get]
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
				"pending_yields":    totalPendingYields,
				"total_redemptions": totalRedemptions,
			},
			"investments":    investments,
			"pending_yields": pendingYields,
			"redemptions":    redemptions,
		},
	})
}

// GetInvestmentHistory returns user's investment history (READ-ONLY)
// @Summary Get investment history
// @Description Get investment history for a wallet address with optional status filtering
// @Tags Portfolio
// @Accept json
// @Produce json
// @Param address path string true "Wallet address"
// @Param status query string false "Investment status filter (active, redeemed)"
// @Success 200 {object} InvestmentListResponse "Investment history"
// @Failure 400 {object} ErrorResponse "Invalid address"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /portfolio/{address}/investments [get]
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
		"data":  investments,
		"count": len(investments),
	})
}

// GetYieldHistory returns user's yield claim history (READ-ONLY)
// @Summary Get yield history
// @Description Get yield claim history for a wallet address with optional status filtering
// @Tags Portfolio
// @Accept json
// @Produce json
// @Param address path string true "Wallet address"
// @Param status query string false "Yield status filter (pending, claimed)"
// @Success 200 {object} YieldClaimListResponse "Yield history"
// @Failure 400 {object} ErrorResponse "Invalid address"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /portfolio/{address}/yields [get]
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
		"data":  yieldClaims,
		"count": len(yieldClaims),
	})
}

// GetPendingYields returns user's pending yield claims (READ-ONLY)
// @Summary Get pending yields
// @Description Get all pending yield claims for a wallet address
// @Tags Portfolio
// @Accept json
// @Produce json
// @Param address path string true "Wallet address"
// @Success 200 {object} YieldClaimListResponse "Pending yields"
// @Failure 400 {object} ErrorResponse "Invalid address"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /portfolio/{address}/yields/pending [get]
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
		"data":  pendingYields,
		"count": len(pendingYields),
	})
}

// GetRedemptionHistory returns user's redemption history (READ-ONLY)
// @Summary Get redemption history
// @Description Get redemption history for a wallet address with optional status filtering
// @Tags Portfolio
// @Accept json
// @Produce json
// @Param address path string true "Wallet address"
// @Param status query string false "Redemption status filter (requested, approved, rejected, completed)"
// @Success 200 {object} RedemptionListResponse "Redemption history"
// @Failure 400 {object} ErrorResponse "Invalid address"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /portfolio/{address}/redemptions [get]
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
		"data":  redemptions,
		"count": len(redemptions),
	})
}
