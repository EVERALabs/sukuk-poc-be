package handlers

import (
	"net/http"

	"sukuk-be/internal/database"
	"sukuk-be/internal/models"
	"sukuk-be/internal/utils"

	"github.com/gin-gonic/gin"
)

// Redemption APIs

// ListRedemptions returns redemption requests and completions
// @Summary List redemptions
// @Description Get a list of redemption requests with optional filtering by investor, sukuk series, and status
// @Tags Redemptions
// @Accept json
// @Produce json
// @Param investor_address query string false "Filter by investor address"
// @Param sukuk_id query string false "Filter by sukuk ID"
// @Param status query string false "Filter by redemption status"
// @Success 200 {object} map[string]interface{} "List of redemptions"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /redemptions [get]
func ListRedemptions(c *gin.Context) {
	var redemptions []models.Redemption

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

	// Order by most recent first
	query = query.Order("created_at DESC")

	if err := query.Find(&redemptions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch redemptions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  redemptions,
		"count": len(redemptions),
		"meta": gin.H{
			"total": len(redemptions),
		},
	})
}

// GetRedemptionsByInvestor returns redemption history for a specific investor
// @Summary Get redemptions by investor
// @Description Get redemption history for a specific investor address
// @Tags Redemptions
// @Accept json
// @Produce json
// @Param investor_address path string true "Investor Ethereum address"
// @Success 200 {object} map[string]interface{} "List of investor's redemptions"
// @Failure 400 {object} map[string]interface{} "Invalid address"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /redemptions/{investor_address} [get]
func GetRedemptionsByInvestor(c *gin.Context) {
	investorAddress := c.Param("investor_address")

	// Validate address format
	if !utils.IsValidEthereumAddress(investorAddress) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Ethereum address format",
		})
		return
	}

	normalizedAddress := utils.NormalizeAddress(investorAddress)

	var redemptions []models.Redemption
	db := database.GetDB()
	if err := db.Preload("Sukuk").Preload("Sukuk.Company").
		Where("investor_address = ?", normalizedAddress).
		Order("created_at DESC").
		Find(&redemptions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch investor redemptions",
		})
		return
	}

	// Calculate redemption summary
	totalRedemptions := len(redemptions)
	completedRedemptions := 0
	pendingRedemptions := 0
	totalRedemptionAmount := "0"

	for _, redemption := range redemptions {
		switch redemption.Status {
		case "completed":
			completedRedemptions++
		case "requested":
			pendingRedemptions++
		}
		// TODO: Sum up redemption amounts properly
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  redemptions,
		"count": totalRedemptions,
		"meta": gin.H{
			"total":                   totalRedemptions,
			"completed":              completedRedemptions,
			"pending":                pendingRedemptions,
			"total_redemption_amount": totalRedemptionAmount,
		},
	})
}

// GetRedemptionsBySukuk returns redemptions for a specific sukuk series
// @Summary Get redemptions by sukuk
// @Description Get all redemption requests for a specific sukuk series
// @Tags Redemptions
// @Accept json
// @Produce json
// @Param sukuk_id path string true "Sukuk ID"
// @Success 200 {object} map[string]interface{} "List of sukuk's redemptions"
// @Failure 400 {object} map[string]interface{} "Invalid sukuk ID"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /redemptions/sukuk/{sukuk_id} [get]
func GetRedemptionsBySukuk(c *gin.Context) {
	sukukID := c.Param("sukuk_id")

	var redemptions []models.Redemption
	db := database.GetDB()
	if err := db.Preload("Sukuk").Preload("Sukuk.Company").
		Where("sukuk_id = ?", sukukID).
		Order("created_at DESC").
		Find(&redemptions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch sukuk redemptions",
		})
		return
	}

	// Calculate sukuk redemption metrics
	totalRedemptions := len(redemptions)
	completedRedemptions := 0
	pendingRedemptions := 0
	uniqueInvestors := make(map[string]bool)

	for _, redemption := range redemptions {
		switch redemption.Status {
		case "completed":
			completedRedemptions++
		case "requested":
			pendingRedemptions++
		}
		uniqueInvestors[redemption.InvestorAddress] = true
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  redemptions,
		"count": totalRedemptions,
		"meta": gin.H{
			"total":             totalRedemptions,
			"completed":        completedRedemptions,
			"pending":          pendingRedemptions,
			"unique_investors": len(uniqueInvestors),
		},
	})
}

// GetRedemptionsByCompany returns redemptions for all sukuk series from a specific company
// @Summary Get redemptions by company
// @Description Get all redemption requests for sukuk series issued by a specific company
// @Tags Redemptions
// @Accept json
// @Produce json
// @Param company_id path string true "Company ID"
// @Success 200 {object} map[string]interface{} "List of company's redemptions"
// @Failure 400 {object} map[string]interface{} "Invalid company ID"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /redemptions/company/{company_id} [get]
func GetRedemptionsByCompany(c *gin.Context) {
	companyID := c.Param("company_id")

	var redemptions []models.Redemption
	db := database.GetDB()
	
	// Join with sukuks to filter by company
	if err := db.Preload("Sukuk").Preload("Sukuk.Company").
		Joins("JOIN sukuks ON redemptions.sukuk_id = sukuks.id").
		Where("sukuks.company_id = ?", companyID).
		Order("redemptions.created_at DESC").
		Find(&redemptions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch company redemptions",
		})
		return
	}

	// Calculate company redemption metrics
	totalRedemptions := len(redemptions)
	completedRedemptions := 0
	pendingRedemptions := 0
	uniqueInvestors := make(map[string]bool)
	uniqueSukuks := make(map[uint]bool)

	for _, redemption := range redemptions {
		switch redemption.Status {
		case "completed":
			completedRedemptions++
		case "requested":
			pendingRedemptions++
		}
		uniqueInvestors[redemption.InvestorAddress] = true
		uniqueSukuks[redemption.SukukID] = true
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  redemptions,
		"count": totalRedemptions,
		"meta": gin.H{
			"total":               totalRedemptions,
			"completed":          completedRedemptions,
			"pending":            pendingRedemptions,
			"unique_investors":   len(uniqueInvestors),
			"unique_sukuks": len(uniqueSukuks),
		},
	})
}

// GetPendingRedemptions returns all pending redemption requests (Admin)
// @Summary Get pending redemptions
// @Description Get all pending redemption requests across all sukuk series (Admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "List of pending redemptions"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security ApiKeyAuth
// @Router /admin/redemptions/pending [get]
func GetPendingRedemptions(c *gin.Context) {
	var redemptions []models.Redemption
	db := database.GetDB()
	
	if err := db.Preload("Sukuk").Preload("Sukuk.Company").
		Where("status = ?", "requested").
		Order("created_at ASC"). // Oldest first for processing
		Find(&redemptions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch pending redemptions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  redemptions,
		"count": len(redemptions),
		"meta": gin.H{
			"total":  len(redemptions),
			"status": "pending",
		},
	})
}