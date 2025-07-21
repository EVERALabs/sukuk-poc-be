package handlers

import (
	"net/http"
	"strings"

	"sukuk-be/internal/database"
	"sukuk-be/internal/models"
	"sukuk-be/internal/utils"

	"github.com/gin-gonic/gin"
)

// Investment APIs

// ListInvestments returns investments with filtering
// @Summary List investments
// @Description Get a list of investments with optional filtering by investor, sukuk series, and status
// @Tags Investments
// @Accept json
// @Produce json
// @Param investor_address query string false "Filter by investor address"
// @Param sukuk_id query string false "Filter by sukuk ID"
// @Param status query string false "Filter by investment status"
// @Success 200 {object} map[string]interface{} "List of investments"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /investments [get]
func ListInvestments(c *gin.Context) {
	var investments []models.Investment

	db := database.GetDB()
	query := db.Preload("Sukuk").Preload("Sukuk.Company")

	// Filter by investor address if provided
	if investorAddress := c.Query("investor_address"); investorAddress != "" {
		// Validate and normalize address
		if !utils.IsValidEthereumAddress(investorAddress) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid Ethereum address format",
			})
			return
		}
		normalizedAddress := utils.NormalizeAddress(investorAddress)
		query = query.Where("investor_address = ?", normalizedAddress)
	}

	// Filter by sukuk if provided
	if sukukID := c.Query("sukuk_id"); sukukID != "" {
		query = query.Where("sukuk_id = ?", sukukID)
	}

	// Filter by status if provided
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&investments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch investments",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  investments,
		"count": len(investments),
		"meta": gin.H{
			"total": len(investments),
		},
	})
}

// GetInvestmentsByInvestor returns all investments for a specific investor
// @Summary Get investments by investor
// @Description Get all investments for a specific investor address
// @Tags Investments
// @Accept json
// @Produce json
// @Param investor_address path string true "Investor Ethereum address"
// @Success 200 {object} map[string]interface{} "List of investor's investments"
// @Failure 400 {object} map[string]interface{} "Invalid address"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /investments/{investor_address} [get]
func GetInvestmentsByInvestor(c *gin.Context) {
	investorAddress := c.Param("investor_address")

	// Validate address format
	if !utils.IsValidEthereumAddress(investorAddress) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Ethereum address format",
		})
		return
	}

	normalizedAddress := utils.NormalizeAddress(investorAddress)

	var investments []models.Investment
	db := database.GetDB()
	if err := db.Preload("Sukuk").Preload("Sukuk.Company").Where("investor_address = ?", normalizedAddress).Find(&investments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch investor investments",
		})
		return
	}

	// Calculate portfolio summary
	activeInvestments := 0
	totalInvestmentValue := "0"
	for _, investment := range investments {
		if investment.Status == "active" {
			activeInvestments++
			// TODO: Add up investment amounts properly
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  investments,
		"count": len(investments),
		"meta": gin.H{
			"total":                 len(investments),
			"active_investments":    activeInvestments,
			"total_investment_value": totalInvestmentValue,
		},
	})
}

// GetInvestmentPortfolio returns portfolio summary for an investor
// @Summary Get investor portfolio summary
// @Description Get portfolio summary including active investments, total value, and performance metrics
// @Tags Investments
// @Accept json
// @Produce json
// @Param investor_address path string true "Investor Ethereum address"
// @Success 200 {object} map[string]interface{} "Portfolio summary"
// @Failure 400 {object} map[string]interface{} "Invalid address"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /investments/{investor_address}/portfolio [get]
func GetInvestmentPortfolio(c *gin.Context) {
	investorAddress := c.Param("investor_address")

	// Validate address format
	if !utils.IsValidEthereumAddress(investorAddress) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Ethereum address format",
		})
		return
	}

	normalizedAddress := utils.NormalizeAddress(investorAddress)
	db := database.GetDB()

	// Get investment summary
	var activeInvestments int64
	var totalInvestments int64
	
	db.Model(&models.Investment{}).Where("investor_address = ?", normalizedAddress).Count(&totalInvestments)
	db.Model(&models.Investment{}).Where("investor_address = ? AND status = ?", normalizedAddress, "active").Count(&activeInvestments)

	// Get yield summary
	var totalYields int64
	var claimedYields int64
	
	db.Model(&models.Yield{}).Where("investor_address = ?", normalizedAddress).Count(&totalYields)
	db.Model(&models.Yield{}).Where("investor_address = ? AND status = ?", normalizedAddress, "claimed").Count(&claimedYields)

	// Get redemption summary
	var totalRedemptions int64
	var completedRedemptions int64
	
	db.Model(&models.Redemption{}).Where("investor_address = ?", normalizedAddress).Count(&totalRedemptions)
	db.Model(&models.Redemption{}).Where("investor_address = ? AND status = ?", normalizedAddress, "completed").Count(&completedRedemptions)

	// Get recent activity (last 5 transactions)
	var recentInvestments []models.Investment
	db.Preload("Sukuk").Where("investor_address = ?", normalizedAddress).Order("created_at DESC").Limit(5).Find(&recentInvestments)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"summary": gin.H{
				"total_investments":     totalInvestments,
				"active_investments":    activeInvestments,
				"total_yields":         totalYields,
				"claimed_yields":       claimedYields,
				"total_redemptions":    totalRedemptions,
				"completed_redemptions": completedRedemptions,
			},
			"recent_activity": recentInvestments,
		},
	})
}

// GetInvestmentsByCompany returns all investments for sukuk series from a specific company
// @Summary Get investments by company
// @Description Get all investments for sukuk series issued by a specific company
// @Tags Investments
// @Accept json
// @Produce json
// @Param company_id path string true "Company ID"
// @Success 200 {object} map[string]interface{} "List of company's investments"
// @Failure 400 {object} map[string]interface{} "Invalid company ID"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /investments/company/{company_id} [get]
func GetInvestmentsByCompany(c *gin.Context) {
	companyID := c.Param("company_id")

	var investments []models.Investment
	db := database.GetDB()
	
	// Join with sukuks to filter by company
	query := db.Preload("Sukuk").Preload("Sukuk.Company").
		Joins("JOIN sukuks ON investments.sukuk_id = sukuks.id").
		Where("sukuks.company_id = ?", companyID)

	if err := query.Find(&investments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch company investments",
		})
		return
	}

	// Calculate company investment summary
	activeInvestments := 0
	uniqueInvestors := make(map[string]bool)
	
	for _, investment := range investments {
		if investment.Status == "active" {
			activeInvestments++
		}
		uniqueInvestors[strings.ToLower(investment.InvestorAddress)] = true
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  investments,
		"count": len(investments),
		"meta": gin.H{
			"total":             len(investments),
			"active":           activeInvestments,
			"unique_investors": len(uniqueInvestors),
		},
	})
}