package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kadzu/sukuk-poc-be/internal/database"
	"github.com/kadzu/sukuk-poc-be/internal/models"
)

func GetYieldClaims(c *gin.Context) {
	var yieldClaims []models.YieldClaim
	
	db := database.GetDB()
	query := db.Preload("SukukSeries").Preload("SukukSeries.Company")

	// Filter by status if provided
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// Filter by investor if provided
	if investor := c.Query("investor"); investor != "" {
		query = query.Where("investor_address = ?", strings.ToLower(investor))
	}

	if err := query.Order("created_at DESC").Find(&yieldClaims).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch yield claims",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": yieldClaims,
		"count": len(yieldClaims),
	})
}

func GetYieldClaim(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid yield claim ID",
		})
		return
	}

	var yieldClaim models.YieldClaim
	db := database.GetDB()
	if err := db.Preload("SukukSeries").Preload("SukukSeries.Company").Preload("Investment").First(&yieldClaim, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Yield claim not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": yieldClaim,
	})
}

func GetYieldClaimsByInvestor(c *gin.Context) {
	address := strings.ToLower(c.Param("address"))
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid investor address",
		})
		return
	}

	var yieldClaims []models.YieldClaim
	db := database.GetDB()
	query := db.Preload("SukukSeries").Preload("SukukSeries.Company").Where("investor_address = ?", address)

	// Filter by status if provided
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&yieldClaims).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch yield claims",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": yieldClaims,
		"count": len(yieldClaims),
	})
}

func GetYieldClaimsBySukuk(c *gin.Context) {
	sukukID, err := strconv.ParseUint(c.Param("sukukId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid sukuk series ID",
		})
		return
	}

	var yieldClaims []models.YieldClaim
	db := database.GetDB()
	query := db.Preload("SukukSeries").Where("sukuk_series_id = ?", uint(sukukID))

	// Filter by status if provided
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&yieldClaims).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch yield claims",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": yieldClaims,
		"count": len(yieldClaims),
	})
}