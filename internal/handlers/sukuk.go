package handlers

import (
	"net/http"
	"strconv"
	"time"

	"sukuk-be/internal/database"
	"sukuk-be/internal/models"
	"sukuk-be/internal/utils"

	"github.com/gin-gonic/gin"
)

// Request/Response Types
type CreateSukukRequest struct {
	CompanyID        uint      `json:"company_id" binding:"required"`
	Name             string    `json:"name" binding:"required"`
	Symbol           string    `json:"symbol" binding:"required"`
	Description      string    `json:"description"`
	TotalSupply      string    `json:"total_supply" binding:"required"`
	YieldRate        float64   `json:"yield_rate" binding:"required"`
	MaturityDate     time.Time `json:"maturity_date" binding:"required"`
	PaymentFrequency int       `json:"payment_frequency"`
	MinInvestment    string    `json:"min_investment" binding:"required"`
	MaxInvestment    string    `json:"max_investment"`
	IsRedeemable     bool      `json:"is_redeemable"`
}

type UpdateSukukRequest struct {
	Name              string `json:"name,omitempty"`
	Description       string `json:"description,omitempty"`
	TokenAddress      string `json:"token_address,omitempty"`
	Status            string `json:"status,omitempty"`
	TotalSupply       string `json:"total_supply,omitempty"`
	OutstandingSupply string `json:"outstanding_supply,omitempty"`
}

// Public Sukuk APIs

// ListSukuk returns a list of all Sukuk series
// @Summary List all Sukuk series
// @Description Get a list of all Sukuk series with optional filtering by company and status
// @Tags Sukuk Series
// @Accept json
// @Produce json
// @Param company_id query string false "Company ID to filter by"
// @Param status query string false "Status to filter by (active, paused, matured)"
// @Success 200 {object} map[string]interface{} "List of Sukuk series"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /sukuks [get]
func ListSukuk(c *gin.Context) {
	var sukuks []models.Sukuk

	db := database.GetDB()
	query := db.Preload("Company")

	// Filter by company if provided
	if companyID := c.Query("company_id"); companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}

	// Filter by status if provided
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&sukuks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch sukuk series",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  sukuks,
		"count": len(sukuks),
		"meta": gin.H{
			"total": len(sukuks),
		},
	})
}

// GetSukuk returns details of a specific Sukuk series
// @Summary Get Sukuk series details
// @Description Get detailed information about a specific Sukuk series including company, investments, yield claims, and redemptions
// @Tags Sukuk Series
// @Accept json
// @Produce json
// @Param id path int true "Sukuk Series ID"
// @Success 200 {object} map[string]interface{} "Sukuk series details"
// @Failure 400 {object} map[string]interface{} "Invalid ID"
// @Failure 404 {object} map[string]interface{} "Sukuk series not found"
// @Router /sukuks/{id} [get]
func GetSukuk(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid sukuk series ID",
		})
		return
	}

	var sukuk models.Sukuk
	db := database.GetDB()
	if err := db.Preload("Company").Preload("Investments").Preload("Yields").Preload("Redemptions").First(&sukuk, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Sukuk series not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": sukuk,
	})
}

// GetSukukMetrics returns performance metrics for a specific Sukuk series
// @Summary Get Sukuk series metrics
// @Description Get performance metrics including total investors, investment amount, pending yields, and redemptions
// @Tags Sukuk Series
// @Accept json
// @Produce json
// @Param id path int true "Sukuk Series ID"
// @Success 200 {object} map[string]interface{} "Sukuk series metrics"
// @Failure 400 {object} map[string]interface{} "Invalid ID"
// @Router /sukuks/{id}/metrics [get]
func GetSukukMetrics(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid sukuk series ID",
		})
		return
	}

	db := database.GetDB()

	// Get total investors
	var totalInvestors int64
	var totalInvestment string = "0"

	db.Model(&models.Investment{}).Where("sukuk_id = ? AND status = ?", uint(id), "active").Count(&totalInvestors)

	// Get pending yields
	var pendingYields int64
	db.Model(&models.Yield{}).Where("sukuk_id = ? AND status = ?", uint(id), "pending").Count(&pendingYields)

	// Get pending redemptions
	var pendingRedemptions int64
	db.Model(&models.Redemption{}).Where("sukuk_id = ? AND status = ?", uint(id), "requested").Count(&pendingRedemptions)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"total_investors":     totalInvestors,
			"total_investment":    totalInvestment,
			"pending_yields":      pendingYields,
			"pending_redemptions": pendingRedemptions,
		},
	})
}

// GetSukukHolders returns current holders of a specific Sukuk
// @Summary Get Sukuk holders
// @Description Get a list of current active investors holding the specified Sukuk series
// @Tags Sukuk Series
// @Accept json
// @Produce json
// @Param id path int true "Sukuk Series ID"
// @Success 200 {object} map[string]interface{} "List of active investments"
// @Failure 400 {object} map[string]interface{} "Invalid ID"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /sukuks/{id}/holders [get]
func GetSukukHolders(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid sukuk series ID",
		})
		return
	}

	var investments []models.Investment
	db := database.GetDB()
	if err := db.Preload("Sukuk").Where("sukuk_id = ? AND status = ?", uint(id), "active").Find(&investments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch sukuk holders",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  investments,
		"count": len(investments),
	})
}

// Admin Sukuk Management APIs

// CreateSukuk creates a new Sukuk series with off-chain data
// @Summary Create new Sukuk series
// @Description Create a new Sukuk series with off-chain data (admin only)
// @Tags Sukuk Management
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body CreateSukukRequest true "Sukuk series data"
// @Success 201 {object} map[string]interface{} "Sukuk series created"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /admin/sukuks [post]
func CreateSukuk(c *gin.Context) {
	var req CreateSukukRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate company exists
	var company models.Company
	db := database.GetDB()
	if err := db.First(&company, req.CompanyID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Company not found",
		})
		return
	}

	// Set defaults
	if req.PaymentFrequency == 0 {
		req.PaymentFrequency = 4 // Quarterly
	}

	sukuk := models.Sukuk{
		CompanyID:         req.CompanyID,
		Name:              req.Name,
		Symbol:            req.Symbol,
		Description:       req.Description,
		TotalSupply:       req.TotalSupply,
		OutstandingSupply: "0",
		YieldRate:         req.YieldRate,
		MaturityDate:      req.MaturityDate,
		PaymentFrequency:  req.PaymentFrequency,
		MinInvestment:     req.MinInvestment,
		MaxInvestment:     req.MaxInvestment,
		Status:            models.SukukStatusActive,
		IsRedeemable:      req.IsRedeemable,
	}

	if err := db.Create(&sukuk).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create sukuk series",
		})
		return
	}

	// Load with company data
	db.Preload("Company").First(&sukuk, sukuk.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Sukuk series created - ready for smart contract deployment",
		"data":    sukuk,
	})
}

// UpdateSukuk updates existing Sukuk series off-chain data
// @Summary Update Sukuk series
// @Description Update existing Sukuk series off-chain data (admin only)
// @Tags Sukuk Management
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Sukuk Series ID"
// @Param request body UpdateSukukRequest true "Update data"
// @Success 200 {object} map[string]interface{} "Sukuk series updated"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Sukuk series not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /admin/sukuks/{id} [put]
func UpdateSukuk(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid sukuk series ID",
		})
		return
	}

	var req UpdateSukukRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var sukuk models.Sukuk
	db := database.GetDB()
	if err := db.First(&sukuk, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Sukuk series not found",
		})
		return
	}

	// Update fields if provided
	if req.Name != "" {
		sukuk.Name = req.Name
	}
	if req.Description != "" {
		sukuk.Description = req.Description
	}
	if req.TokenAddress != "" {
		sukuk.TokenAddress = req.TokenAddress
	}
	if req.Status != "" {
		sukuk.Status = models.SukukStatus(req.Status)
	}
	if req.TotalSupply != "" {
		sukuk.TotalSupply = req.TotalSupply
	}
	if req.OutstandingSupply != "" {
		sukuk.OutstandingSupply = req.OutstandingSupply
	}

	if err := db.Save(&sukuk).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update sukuk series",
		})
		return
	}

	// Load with company data
	db.Preload("Company").First(&sukuk, sukuk.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Sukuk series updated",
		"data":    sukuk,
	})
}

// UploadProspectus handles PDF prospectus file upload
// @Summary Upload Sukuk prospectus
// @Description Upload PDF prospectus file for a Sukuk series (admin only)
// @Tags Sukuk Management
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Sukuk Series ID"
// @Param file formData file true "PDF prospectus file"
// @Success 200 {object} map[string]interface{} "Prospectus uploaded"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Sukuk series not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /admin/sukuks/{id}/upload-prospectus [post]
func UploadProspectus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid sukuk series ID",
		})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}

	// Check if sukuk series exists
	var sukuk models.Sukuk
	db := database.GetDB()
	if err := db.First(&sukuk, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Sukuk series not found",
		})
		return
	}

	// Configure file upload
	config := utils.DefaultPDFConfig("./uploads/prospectus")

	// Save file with validation
	filename, url, err := utils.SaveFile(file, config, strconv.FormatUint(id, 10))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Delete old prospectus if exists
	if sukuk.Prospectus != "" {
		utils.DeleteFile(sukuk.Prospectus)
	}

	// Update sukuk series prospectus URL
	sukuk.Prospectus = url
	if err := db.Save(&sukuk).Error; err != nil {
		// Clean up uploaded file if database update fails
		utils.DeleteFile(url)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update prospectus URL",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Prospectus uploaded successfully",
		"filename": filename,
		"url":      sukuk.Prospectus,
	})
}