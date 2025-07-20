package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"sukuk-be/internal/database"
	"sukuk-be/internal/models"

	"github.com/gin-gonic/gin"
)

// CreateRedemptionRequest represents the request body for creating a redemption request
type CreateRedemptionRequest struct {
	SukukSeriesID    uint   `json:"sukuk_series_id" binding:"required" swaggertype:"integer" example:"1"`
	InvestmentID     uint   `json:"investment_id" binding:"required" swaggertype:"integer" example:"1"`
	InvestorAddress  string `json:"investor_address" binding:"required" swaggertype:"string" example:"0x1234567890123456789012345678901234567890"`
	TokenAmount      string `json:"token_amount" binding:"required" swaggertype:"string" example:"500000000000000000000"`
	RedemptionAmount string `json:"redemption_amount" binding:"required" swaggertype:"string" example:"500000000000000000000"`
	RequestReason    string `json:"request_reason" swaggertype:"string" example:"Early redemption for emergency needs"`
}

// ApproveRedemptionRequest represents the request body for approving a redemption
type ApproveRedemptionRequest struct {
	ApprovalNotes string `json:"approval_notes" swaggertype:"string" example:"Redemption approved after verification"`
}

// RejectRedemptionRequest represents the request body for rejecting a redemption
type RejectRedemptionRequest struct {
	RejectionReason string `json:"rejection_reason" binding:"required" swaggertype:"string" example:"Insufficient documentation provided"`
}

// GetRedemptions returns a list of all redemptions with optional filtering
// @Summary List all redemptions
// @Description Get a list of all redemptions with optional filtering by status and investor address
// @Tags Redemptions
// @Accept json
// @Produce json
// @Param status query string false "Redemption status to filter by (requested, approved, rejected, completed)"
// @Param investor query string false "Investor wallet address to filter by"
// @Success 200 {object} RedemptionListResponse "List of redemptions"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /redemptions [get]
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
		"data":  redemptions,
		"count": len(redemptions),
	})
}

// GetRedemption returns details of a specific redemption
// @Summary Get redemption details
// @Description Get detailed information about a specific redemption including Sukuk series, company, and investment data
// @Tags Redemptions
// @Accept json
// @Produce json
// @Param id path int true "Redemption ID"
// @Success 200 {object} RedemptionResponse "Redemption details"
// @Failure 400 {object} ErrorResponse "Invalid ID"
// @Failure 404 {object} ErrorResponse "Redemption not found"
// @Router /redemptions/{id} [get]
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

// GetRedemptionsByInvestor returns all redemptions for a specific investor
// @Summary Get redemptions by investor
// @Description Get all redemptions made by a specific wallet address with optional status filtering
// @Tags Redemptions
// @Accept json
// @Produce json
// @Param address path string true "Investor wallet address"
// @Param status query string false "Redemption status to filter by (requested, approved, rejected, completed)"
// @Success 200 {object} RedemptionListResponse "Investor's redemptions"
// @Failure 400 {object} ErrorResponse "Invalid address"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /redemptions/investor/{address} [get]
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
		"data":  redemptions,
		"count": len(redemptions),
	})
}

// GetRedemptionsBySukuk returns all redemptions for a specific Sukuk series
// @Summary Get redemptions by Sukuk series
// @Description Get all redemptions for a specific Sukuk series with optional status filtering
// @Tags Redemptions
// @Accept json
// @Produce json
// @Param sukukId path int true "Sukuk Series ID"
// @Param status query string false "Redemption status to filter by (requested, approved, rejected, completed)"
// @Success 200 {object} RedemptionListResponse "Sukuk series redemptions"
// @Failure 400 {object} ErrorResponse "Invalid ID"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /redemptions/sukuk/{sukukId} [get]
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
		"data":  redemptions,
		"count": len(redemptions),
	})
}

// CreateRedemption creates a new redemption request
// @Summary Create redemption request
// @Description Create a new redemption request for an active investment
// @Tags Redemptions
// @Accept json
// @Produce json
// @Param request body CreateRedemptionRequest true "Redemption request data"
// @Success 201 {object} APIResponse "Redemption request created"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /redemptions [post]
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

// ApproveRedemption approves a redemption request (admin only)
// @Summary Approve redemption
// @Description Approve a redemption request (admin only)
// @Tags Redemptions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Redemption ID"
// @Param request body ApproveRedemptionRequest true "Approval data"
// @Success 200 {object} APIResponse "Redemption approved"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 404 {object} ErrorResponse "Redemption not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /redemptions/{id}/approve [put]
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

// RejectRedemption rejects a redemption request (admin only)
// @Summary Reject redemption
// @Description Reject a redemption request (admin only)
// @Tags Redemptions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Redemption ID"
// @Param request body RejectRedemptionRequest true "Rejection data"
// @Success 200 {object} APIResponse "Redemption rejected"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 404 {object} ErrorResponse "Redemption not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /redemptions/{id}/reject [put]
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
