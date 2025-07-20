package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kadzu/sukuk-poc-be/internal/database"
	"github.com/kadzu/sukuk-poc-be/internal/models"
)

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