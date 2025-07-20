package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kadzu/sukuk-poc-be/internal/database"
	"github.com/kadzu/sukuk-poc-be/internal/models"
)

// GetInvestments returns a list of all investments with optional filtering
// @Summary List all investments
// @Description Get a list of all investments with optional filtering by investor address and status
// @Tags Investments
// @Accept json
// @Produce json
// @Param investor_address query string false "Investor wallet address to filter by"
// @Param status query string false "Investment status to filter by (active, redeemed)"
// @Success 200 {object} InvestmentListResponse "List of investments"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /investments [get]
func GetInvestments(c *gin.Context) {
	var investments []models.Investment
	
	db := database.GetDB()
	query := db.Preload("SukukSeries").Preload("SukukSeries.Company")

	// Filter by investor address if provided
	if investorAddress := c.Query("investor_address"); investorAddress != "" {
		query = query.Where("investor_address = ?", investorAddress)
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
		"data": investments,
		"count": len(investments),
	})
}

// GetInvestment returns details of a specific investment
// @Summary Get investment details
// @Description Get detailed information about a specific investment including Sukuk series and company data
// @Tags Investments
// @Accept json
// @Produce json
// @Param id path int true "Investment ID"
// @Success 200 {object} InvestmentResponse "Investment details"
// @Failure 400 {object} ErrorResponse "Invalid ID"
// @Failure 404 {object} ErrorResponse "Investment not found"
// @Router /investments/{id} [get]
func GetInvestment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid investment ID",
		})
		return
	}

	var investment models.Investment
	db := database.GetDB()
	if err := db.Preload("SukukSeries").Preload("SukukSeries.Company").First(&investment, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Investment not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": investment,
	})
}

// GetInvestmentsByInvestor returns all investments for a specific investor
// @Summary Get investments by investor
// @Description Get all investments made by a specific wallet address
// @Tags Investments
// @Accept json
// @Produce json
// @Param address path string true "Investor wallet address"
// @Success 200 {object} InvestmentListResponse "Investor's investments"
// @Failure 400 {object} ErrorResponse "Invalid address"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /investments/investor/{address} [get]
func GetInvestmentsByInvestor(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid investor address",
		})
		return
	}

	var investments []models.Investment
	db := database.GetDB()
	if err := db.Preload("SukukSeries").Preload("SukukSeries.Company").Where("investor_address = ?", address).Find(&investments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch investments",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": investments,
		"count": len(investments),
	})
}

// GetInvestmentsBySukuk returns all investments for a specific Sukuk series
// @Summary Get investments by Sukuk series
// @Description Get all investments made in a specific Sukuk series
// @Tags Investments
// @Accept json
// @Produce json
// @Param sukukId path int true "Sukuk Series ID"
// @Success 200 {object} InvestmentListResponse "Sukuk series investments"
// @Failure 400 {object} ErrorResponse "Invalid ID"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /investments/sukuk/{sukukId} [get]
func GetInvestmentsBySukuk(c *gin.Context) {
	sukukID, err := strconv.ParseUint(c.Param("sukukId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid sukuk series ID",
		})
		return
	}

	var investments []models.Investment
	db := database.GetDB()
	if err := db.Preload("SukukSeries").Where("sukuk_series_id = ?", uint(sukukID)).Find(&investments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch investments",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": investments,
		"count": len(investments),
	})
}