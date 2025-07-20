package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/kadzu/sukuk-poc-be/internal/database"
	"github.com/kadzu/sukuk-poc-be/internal/models"
	"github.com/kadzu/sukuk-poc-be/internal/utils"
)

// CreateSukukSeriesRequest represents the request body for creating a new Sukuk series
type CreateSukukSeriesRequest struct {
	CompanyID              uint      `json:"company_id" binding:"required" swaggertype:"integer" example:"1"`
	Name                   string    `json:"name" binding:"required" swaggertype:"string" example:"Green Sukuk Series A"`
	Symbol                 string    `json:"symbol" binding:"required" swaggertype:"string" example:"GSA"`
	Description            string    `json:"description" swaggertype:"string" example:"Sustainable infrastructure financing sukuk"`
	TotalSupply            string    `json:"total_supply" binding:"required" swaggertype:"string" example:"1000000000000000000000000"`
	YieldRate              float64   `json:"yield_rate" binding:"required" swaggertype:"number" example:"0.085"`
	MaturityDate           time.Time `json:"maturity_date" binding:"required" swaggertype:"string" example:"2027-12-31T00:00:00Z"`
	PaymentFrequency       int       `json:"payment_frequency" swaggertype:"integer" example:"4"`
	MinInvestment          string    `json:"min_investment" binding:"required" swaggertype:"string" example:"1000000000000000000"`
	MaxInvestment          string    `json:"max_investment" swaggertype:"string" example:"100000000000000000000"`
	IsRedeemable           bool      `json:"is_redeemable" swaggertype:"boolean" example:"true"`
}

// UpdateSukukSeriesRequest represents the request body for updating a Sukuk series
type UpdateSukukSeriesRequest struct {
	Name                   string    `json:"name,omitempty" swaggertype:"string" example:"Updated Green Sukuk Series A"`
	Description            string    `json:"description,omitempty" swaggertype:"string" example:"Updated description"`
	TokenAddress           string    `json:"token_address,omitempty" swaggertype:"string" example:"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"`
	Status                 string    `json:"status,omitempty" swaggertype:"string" example:"active"`
	TotalSupply            string    `json:"total_supply,omitempty" swaggertype:"string" example:"2000000000000000000000000"`
	OutstandingSupply      string    `json:"outstanding_supply,omitempty" swaggertype:"string" example:"1000000000000000000000000"`
}

// CreateSukukSeries creates a new Sukuk series with off-chain data
// @Summary Create new Sukuk series
// @Description Create a new Sukuk series with off-chain data (admin only)
// @Tags Sukuk Management
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body CreateSukukSeriesRequest true "Sukuk series data"
// @Success 201 {object} APIResponse "Sukuk series created"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /admin/sukuks [post]
func CreateSukukSeries(c *gin.Context) {
	var req CreateSukukSeriesRequest
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

	sukukSeries := models.SukukSeries{
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

	if err := db.Create(&sukukSeries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create sukuk series",
		})
		return
	}

	// Load with company data
	db.Preload("Company").First(&sukukSeries, sukukSeries.ID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Sukuk series created - ready for smart contract deployment",
		"data":    sukukSeries,
	})
}

// UpdateSukukSeries updates existing Sukuk series off-chain data
// @Summary Update Sukuk series
// @Description Update existing Sukuk series off-chain data (admin only)
// @Tags Sukuk Management
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Sukuk Series ID"
// @Param request body UpdateSukukSeriesRequest true "Update data"
// @Success 200 {object} APIResponse "Sukuk series updated"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 404 {object} ErrorResponse "Sukuk series not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /admin/sukuks/{id} [put]
func UpdateSukukSeries(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid sukuk series ID",
		})
		return
	}
	
	var req UpdateSukukSeriesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var sukukSeries models.SukukSeries
	db := database.GetDB()
	if err := db.First(&sukukSeries, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Sukuk series not found",
		})
		return
	}

	// Update fields if provided
	if req.Name != "" {
		sukukSeries.Name = req.Name
	}
	if req.Description != "" {
		sukukSeries.Description = req.Description
	}
	if req.TokenAddress != "" {
		sukukSeries.TokenAddress = req.TokenAddress
	}
	if req.Status != "" {
		sukukSeries.Status = models.SukukStatus(req.Status)
	}
	if req.TotalSupply != "" {
		sukukSeries.TotalSupply = req.TotalSupply
	}
	if req.OutstandingSupply != "" {
		sukukSeries.OutstandingSupply = req.OutstandingSupply
	}

	if err := db.Save(&sukukSeries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update sukuk series",
		})
		return
	}

	// Load with company data
	db.Preload("Company").First(&sukukSeries, sukukSeries.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Sukuk series updated",
		"data":    sukukSeries,
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
// @Success 200 {object} FileUploadResponse "Prospectus uploaded"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 404 {object} ErrorResponse "Sukuk series not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
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
	var sukukSeries models.SukukSeries
	db := database.GetDB()
	if err := db.First(&sukukSeries, uint(id)).Error; err != nil {
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
	if sukukSeries.Prospectus != "" {
		utils.DeleteFile(sukukSeries.Prospectus)
	}

	// Update sukuk series prospectus URL
	sukukSeries.Prospectus = url
	if err := db.Save(&sukukSeries).Error; err != nil {
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
		"url":      sukukSeries.Prospectus,
	})
}