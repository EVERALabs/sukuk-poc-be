package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kadzu/sukuk-poc-be/internal/database"
	"github.com/kadzu/sukuk-poc-be/internal/models"
)

type CreateRedemptionRequest struct {
	SukukSeriesID   uint   `json:"sukuk_series_id" binding:"required"`
	InvestmentID    uint   `json:"investment_id" binding:"required"`
	InvestorAddress string `json:"investor_address" binding:"required"`
	TokenAmount     string `json:"token_amount" binding:"required"`
	RedemptionAmount string `json:"redemption_amount" binding:"required"`
	RequestReason   string `json:"request_reason"`
}

type ApproveRedemptionRequest struct {
	ApprovalNotes string `json:"approval_notes"`
}

type RejectRedemptionRequest struct {
	RejectionReason string `json:"rejection_reason" binding:"required"`
}

func GetRedemptions(c *gin.Context) {
	var redemptions []models.Redemption
	
	db := database.GetDB()
	query := db.Preload("SukukSeries").Preload("SukukSeries.Company").Preload("Investment")

	// Filter by status if provided
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// Filter by investor if provided
	if investor := c.Query("investor"); investor != "" {
		query = query.Where("investor_address = ?", strings.ToLower(investor))
	}

	if err := query.Order("created_at DESC").Find(&redemptions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch redemptions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": redemptions,
		"count": len(redemptions),
	})
}

func GetRedemption(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid redemption ID",
		})
		return
	}

	var redemption models.Redemption
	db := database.GetDB()
	if err := db.Preload("SukukSeries").Preload("SukukSeries.Company").Preload("Investment").First(&redemption, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Redemption not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": redemption,
	})
}

func GetRedemptionsByInvestor(c *gin.Context) {
	address := strings.ToLower(c.Param("address"))
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid investor address",
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
			"error": "Failed to fetch redemptions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": redemptions,
		"count": len(redemptions),
	})
}

func GetRedemptionsBySukuk(c *gin.Context) {
	sukukID, err := strconv.ParseUint(c.Param("sukukId"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid sukuk series ID",
		})
		return
	}

	var redemptions []models.Redemption
	db := database.GetDB()
	query := db.Preload("SukukSeries").Preload("Investment").Where("sukuk_series_id = ?", uint(sukukID))

	// Filter by status if provided
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&redemptions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch redemptions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": redemptions,
		"count": len(redemptions),
	})
}

func CreateRedemption(c *gin.Context) {
	var req CreateRedemptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate investment exists and belongs to investor
	var investment models.Investment
	db := database.GetDB()
	if err := db.Where("id = ? AND investor_address = ? AND status = ?", 
		req.InvestmentID, strings.ToLower(req.InvestorAddress), "active").First(&investment).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Investment not found or not active",
		})
		return
	}

	redemption := models.Redemption{
		SukukSeriesID:    req.SukukSeriesID,
		InvestmentID:     req.InvestmentID,
		InvestorAddress:  strings.ToLower(req.InvestorAddress),
		TokenAmount:      req.TokenAmount,
		RedemptionAmount: req.RedemptionAmount,
		RequestReason:    req.RequestReason,
		Status:           models.RedemptionStatusRequested,
		RequestedAt:      time.Now(),
	}

	if err := db.Create(&redemption).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create redemption request",
		})
		return
	}

	// Load with relations
	db.Preload("SukukSeries").Preload("Investment").First(&redemption, redemption.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Redemption request created successfully",
		"data":    redemption,
	})
}

func ApproveRedemption(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid redemption ID",
		})
		return
	}

	var req ApproveRedemptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var redemption models.Redemption
	db := database.GetDB()
	if err := db.First(&redemption, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Redemption not found",
		})
		return
	}

	if redemption.Status != models.RedemptionStatusRequested {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Redemption is not in requested status",
		})
		return
	}

	now := time.Now()
	redemption.Status = models.RedemptionStatusApproved
	redemption.ApprovedAt = &now
	redemption.ApprovalNotes = req.ApprovalNotes

	if err := db.Save(&redemption).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to approve redemption",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Redemption approved successfully",
		"data":    redemption,
	})
}

func RejectRedemption(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid redemption ID",
		})
		return
	}

	var req RejectRedemptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var redemption models.Redemption
	db := database.GetDB()
	if err := db.First(&redemption, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Redemption not found",
		})
		return
	}

	if redemption.Status != models.RedemptionStatusRequested {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Redemption is not in requested status",
		})
		return
	}

	now := time.Now()
	redemption.Status = models.RedemptionStatusRejected
	redemption.RejectedAt = &now
	redemption.RejectionReason = req.RejectionReason

	if err := db.Save(&redemption).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to reject redemption",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Redemption rejected successfully",
		"data":    redemption,
	})
}