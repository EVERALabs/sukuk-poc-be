package handlers

import (
	"net/http"

	"sukuk-be/internal/database"
	"sukuk-be/internal/models"
	"sukuk-be/internal/utils"

	"github.com/gin-gonic/gin"
)

// Yield APIs

// ListYields returns yield distributions and claims
// @Summary List yields
// @Description Get a list of yield distributions and claims with optional filtering
// @Tags Yields
// @Accept json
// @Produce json
// @Param investor_address query string false "Filter by investor address"
// @Param sukuk_id query string false "Filter by sukuk ID"
// @Param status query string false "Filter by claim status (pending, claimed)"
// @Success 200 {object} map[string]interface{} "List of yields"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /yields [get]
func ListYields(c *gin.Context) {
	var yields []models.Yield

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

	// Order by most recent distribution first
	query = query.Order("distribution_date DESC")

	if err := query.Find(&yields).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch yields",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  yields,
		"count": len(yields),
		"meta": gin.H{
			"total": len(yields),
		},
	})
}

// GetYieldsByInvestor returns yield history for a specific investor
// @Summary Get yields by investor
// @Description Get yield distribution and claim history for a specific investor address
// @Tags Yields
// @Accept json
// @Produce json
// @Param investor_address path string true "Investor Ethereum address"
// @Success 200 {object} map[string]interface{} "List of investor's yields"
// @Failure 400 {object} map[string]interface{} "Invalid address"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /yields/{investor_address} [get]
func GetYieldsByInvestor(c *gin.Context) {
	investorAddress := c.Param("investor_address")

	// Validate address format
	if !utils.IsValidEthereumAddress(investorAddress) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Ethereum address format",
		})
		return
	}

	normalizedAddress := utils.NormalizeAddress(investorAddress)

	var yields []models.Yield
	db := database.GetDB()
	if err := db.Preload("Sukuk").Preload("Sukuk.Company").
		Where("investor_address = ?", normalizedAddress).
		Order("distribution_date DESC").
		Find(&yields).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch investor yields",
		})
		return
	}

	// Calculate yield summary
	totalYields := len(yields)
	claimedYields := 0
	pendingYields := 0
	totalYieldAmount := "0"
	totalClaimedAmount := "0"

	for _, yield := range yields {
		switch yield.Status {
		case "claimed":
			claimedYields++
		case "pending":
			pendingYields++
		}
		// TODO: Sum up yield amounts properly
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  yields,
		"count": totalYields,
		"meta": gin.H{
			"total":                totalYields,
			"claimed":             claimedYields,
			"pending":             pendingYields,
			"total_yield_amount":   totalYieldAmount,
			"total_claimed_amount": totalClaimedAmount,
		},
	})
}

// GetYieldsBySukuk returns yield distributions for a specific sukuk series
// @Summary Get yields by sukuk
// @Description Get all yield distributions for a specific sukuk series
// @Tags Yields
// @Accept json
// @Produce json
// @Param sukuk_id path string true "Sukuk ID"
// @Success 200 {object} map[string]interface{} "List of sukuk's yields"
// @Failure 400 {object} map[string]interface{} "Invalid sukuk ID"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /yields/sukuk/{sukuk_id} [get]
func GetYieldsBySukuk(c *gin.Context) {
	sukukID := c.Param("sukuk_id")

	var yields []models.Yield
	db := database.GetDB()
	if err := db.Preload("Sukuk").Preload("Sukuk.Company").
		Where("sukuk_id = ?", sukukID).
		Order("distribution_date DESC").
		Find(&yields).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch sukuk yields",
		})
		return
	}

	// Calculate sukuk yield metrics
	totalYields := len(yields)
	claimedYields := 0
	pendingYields := 0
	uniqueInvestors := make(map[string]bool)
	distributionDates := make(map[string]bool)

	for _, yield := range yields {
		switch yield.Status {
		case "claimed":
			claimedYields++
		case "pending":
			pendingYields++
		}
		uniqueInvestors[yield.InvestorAddress] = true
		distributionDates[yield.DistributionDate.Format("2006-01-02")] = true
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  yields,
		"count": totalYields,
		"meta": gin.H{
			"total":               totalYields,
			"claimed":            claimedYields,
			"pending":            pendingYields,
			"unique_investors":   len(uniqueInvestors),
			"distribution_dates": len(distributionDates),
		},
	})
}

// GetYieldsByCompany returns yields for all sukuk series from a specific company
// @Summary Get yields by company
// @Description Get all yield distributions for sukuk series issued by a specific company
// @Tags Yields
// @Accept json
// @Produce json
// @Param company_id path string true "Company ID"
// @Success 200 {object} map[string]interface{} "List of company's yields"
// @Failure 400 {object} map[string]interface{} "Invalid company ID"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /yields/company/{company_id} [get]
func GetYieldsByCompany(c *gin.Context) {
	companyID := c.Param("company_id")

	var yields []models.Yield
	db := database.GetDB()
	
	// Join with sukuks to filter by company
	if err := db.Preload("Sukuk").Preload("Sukuk.Company").
		Joins("JOIN sukuks ON yields.sukuk_id = sukuks.id").
		Where("sukuks.company_id = ?", companyID).
		Order("yields.distribution_date DESC").
		Find(&yields).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch company yields",
		})
		return
	}

	// Calculate company yield metrics
	totalYields := len(yields)
	claimedYields := 0
	pendingYields := 0
	uniqueInvestors := make(map[string]bool)
	uniqueSukuks := make(map[uint]bool)

	for _, yield := range yields {
		switch yield.Status {
		case "claimed":
			claimedYields++
		case "pending":
			pendingYields++
		}
		uniqueInvestors[yield.InvestorAddress] = true
		uniqueSukuks[yield.SukukID] = true
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  yields,
		"count": totalYields,
		"meta": gin.H{
			"total":               totalYields,
			"claimed":            claimedYields,
			"pending":            pendingYields,
			"unique_investors":   len(uniqueInvestors),
			"unique_sukuks": len(uniqueSukuks),
		},
	})
}

// GetPendingYields returns all pending yield claims (Admin)
// @Summary Get pending yields
// @Description Get all pending yield claims across all sukuk series (Admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "List of pending yields"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security ApiKeyAuth
// @Router /admin/yields/pending [get]
func GetPendingYields(c *gin.Context) {
	var yields []models.Yield
	db := database.GetDB()
	
	if err := db.Preload("Sukuk").Preload("Sukuk.Company").
		Where("status = ?", "pending").
		Order("distribution_date ASC"). // Oldest distributions first
		Find(&yields).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch pending yields",
		})
		return
	}

	// Group by distribution date for admin overview
	distributionGroups := make(map[string][]models.Yield)
	for _, yield := range yields {
		dateKey := yield.DistributionDate.Format("2006-01-02")
		distributionGroups[dateKey] = append(distributionGroups[dateKey], yield)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  yields,
		"count": len(yields),
		"meta": gin.H{
			"total":               len(yields),
			"status":             "pending",
			"distribution_groups": len(distributionGroups),
		},
	})
}

// GetYieldDistributions returns yield distribution summary by date (Admin)
// @Summary Get yield distributions
// @Description Get yield distribution summary grouped by distribution date (Admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Yield distribution summary"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security ApiKeyAuth
// @Router /admin/yields/distributions [get]
func GetYieldDistributions(c *gin.Context) {
	var yields []models.Yield
	db := database.GetDB()
	
	if err := db.Preload("Sukuk").Preload("Sukuk.Company").
		Order("distribution_date DESC").
		Find(&yields).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch yield distributions",
		})
		return
	}

	// Group yields by distribution date and sukuk series
	type DistributionSummary struct {
		Date            string `json:"date"`
		SukukID   uint   `json:"sukuk_id"`
		SukukName string `json:"sukuk_name"`
		CompanyName     string `json:"company_name"`
		TotalClaims     int    `json:"total_claims"`
		ClaimedCount    int    `json:"claimed_count"`
		PendingCount    int    `json:"pending_count"`
	}

	summaryMap := make(map[string]*DistributionSummary)
	
	for _, yield := range yields {
		key := yield.DistributionDate.Format("2006-01-02") + "_" + string(rune(yield.SukukID))
		
		if summary, exists := summaryMap[key]; exists {
			summary.TotalClaims++
			if yield.Status == "claimed" {
				summary.ClaimedCount++
			} else {
				summary.PendingCount++
			}
		} else {
			summary := &DistributionSummary{
				Date:            yield.DistributionDate.Format("2006-01-02"),
				SukukID:   yield.SukukID,
				SukukName: yield.Sukuk.Name,
				CompanyName:     yield.Sukuk.Company.Name,
				TotalClaims:     1,
			}
			if yield.Status == "claimed" {
				summary.ClaimedCount = 1
			} else {
				summary.PendingCount = 1
			}
			summaryMap[key] = summary
		}
	}

	// Convert map to slice
	distributions := make([]DistributionSummary, 0, len(summaryMap))
	for _, summary := range summaryMap {
		distributions = append(distributions, *summary)
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  distributions,
		"count": len(distributions),
		"meta": gin.H{
			"total_distributions": len(distributions),
			"total_claims":       len(yields),
		},
	})
}