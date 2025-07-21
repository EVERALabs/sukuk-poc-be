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
// @Param page query int false "Page number (default: 1)"
// @Param per_page query int false "Items per page (default: 20)"
// @Success 200 {object} SukukResponse "List of Sukuk series"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /sukuks [get]
func ListSukuk(c *gin.Context) {
	var sukuks []models.Sukuk
	var total int64

	db := database.GetDB()
	query := db.Model(&models.Sukuk{}).Preload("Company")

	// Filter by company if provided
	if companyID := c.Query("company_id"); companyID != "" {
		query = query.Where("company_id = ?", companyID)
	}

	// Filter by status if provided
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// Count total records before pagination
	query.Count(&total)

	// Pagination
	page := 1
	perPage := 20
	if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && p > 0 {
		page = p
	}
	if pp, err := strconv.Atoi(c.DefaultQuery("per_page", "20")); err == nil && pp > 0 {
		perPage = pp
	}

	offset := (page - 1) * perPage
	query = query.Offset(offset).Limit(perPage)

	if err := query.Find(&sukuks).Error; err != nil {
		InternalServerError(c, "Failed to fetch sukuk series")
		return
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	pagination := &Pagination{
		Total:       int(total),
		Count:       len(sukuks),
		Page:        page,
		PerPage:     perPage,
		TotalPages:  totalPages,
		HasNext:     page < totalPages,
		HasPrevious: page > 1,
	}

	SendPaginatedResponse(c, sukuks, pagination)
}

// GetSukuk returns details of a specific Sukuk series
// @Summary Get Sukuk series details
// @Description Get detailed information about a specific Sukuk series including company, investments, yield claims, and redemptions
// @Tags Sukuk Series
// @Accept json
// @Produce json
// @Param id path int true "Sukuk Series ID"
// @Success 200 {object} SukukResponse "Sukuk series details"
// @Failure 400 {object} APIResponse "Invalid ID"
// @Failure 404 {object} APIResponse "Sukuk series not found"
// @Router /sukuks/{id} [get]
func GetSukuk(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid sukuk series ID")
		return
	}

	var sukuk models.Sukuk
	db := database.GetDB()
	if err := db.Preload("Company").Preload("Investments").Preload("Yields").Preload("Redemptions").First(&sukuk, uint(id)).Error; err != nil {
		NotFound(c, "Sukuk series not found")
		return
	}

	SendSuccess(c, http.StatusOK, sukuk, "")
}

// GetSukukMetrics returns performance metrics for a specific Sukuk series
// @Summary Get Sukuk series metrics
// @Description Get performance metrics including total investors, investment amount, pending yields, and redemptions
// @Tags Sukuk Series
// @Accept json
// @Produce json
// @Param id path int true "Sukuk Series ID"
// @Success 200 {object} SystemResponse "Sukuk series metrics"
// @Failure 400 {object} APIResponse "Invalid ID"
// @Failure 404 {object} APIResponse "Sukuk series not found"
// @Router /sukuks/{id}/metrics [get]
func GetSukukMetrics(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid sukuk series ID")
		return
	}

	db := database.GetDB()

	// Check if sukuk exists
	var exists bool
	db.Model(&models.Sukuk{}).Where("id = ?", uint(id)).Select("1").Scan(&exists)
	if !exists {
		NotFound(c, "Sukuk series not found")
		return
	}

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

	metrics := gin.H{
		"total_investors":     totalInvestors,
		"total_investment":    totalInvestment,
		"pending_yields":      pendingYields,
		"pending_redemptions": pendingRedemptions,
	}

	SendSuccess(c, http.StatusOK, metrics, "")
}

// GetSukukHolders returns current holders of a specific Sukuk
// @Summary Get Sukuk holders
// @Description Get a list of current active investors holding the specified Sukuk series
// @Tags Sukuk Series
// @Accept json
// @Produce json
// @Param id path int true "Sukuk Series ID"
// @Param page query int false "Page number (default: 1)"
// @Param per_page query int false "Items per page (default: 20)"
// @Success 200 {object} InvestmentResponse "List of active investments"
// @Failure 400 {object} APIResponse "Invalid ID"
// @Failure 404 {object} APIResponse "Sukuk series not found"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /sukuks/{id}/holders [get]
func GetSukukHolders(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid sukuk series ID")
		return
	}

	// Check if sukuk exists
	db := database.GetDB()
	var exists bool
	db.Model(&models.Sukuk{}).Where("id = ?", uint(id)).Select("1").Scan(&exists)
	if !exists {
		NotFound(c, "Sukuk series not found")
		return
	}

	var investments []models.Investment
	var total int64

	query := db.Model(&models.Investment{}).Preload("Sukuk").Where("sukuk_id = ? AND status = ?", uint(id), "active")

	// Count total records before pagination
	query.Count(&total)

	// Pagination
	page := 1
	perPage := 20
	if p, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && p > 0 {
		page = p
	}
	if pp, err := strconv.Atoi(c.DefaultQuery("per_page", "20")); err == nil && pp > 0 {
		perPage = pp
	}

	offset := (page - 1) * perPage
	query = query.Offset(offset).Limit(perPage)

	if err := query.Find(&investments).Error; err != nil {
		InternalServerError(c, "Failed to fetch sukuk holders")
		return
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	pagination := &Pagination{
		Total:       int(total),
		Count:       len(investments),
		Page:        page,
		PerPage:     perPage,
		TotalPages:  totalPages,
		HasNext:     page < totalPages,
		HasPrevious: page > 1,
	}

	SendPaginatedResponse(c, investments, pagination)
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
// @Success 201 {object} SukukResponse "Sukuk series created"
// @Failure 400 {object} APIResponse "Bad request"
// @Failure 422 {object} APIResponse "Validation error"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /admin/sukuks [post]
func CreateSukuk(c *gin.Context) {
	var req CreateSukukRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, err.Error())
		return
	}

	// Validate company exists
	var company models.Company
	db := database.GetDB()
	if err := db.First(&company, req.CompanyID).Error; err != nil {
		BadRequest(c, "Company not found")
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
		InternalServerError(c, "Failed to create sukuk series")
		return
	}

	// Load with company data
	db.Preload("Company").First(&sukuk, sukuk.ID)

	SendSuccess(c, http.StatusCreated, sukuk, "Sukuk series created - ready for smart contract deployment")
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
// @Success 200 {object} SukukResponse "Sukuk series updated"
// @Failure 400 {object} APIResponse "Bad request"
// @Failure 404 {object} APIResponse "Sukuk series not found"
// @Failure 422 {object} APIResponse "Validation error"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /admin/sukuks/{id} [put]
func UpdateSukuk(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid sukuk series ID")
		return
	}

	var req UpdateSukukRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationError(c, err.Error())
		return
	}

	var sukuk models.Sukuk
	db := database.GetDB()
	if err := db.First(&sukuk, uint(id)).Error; err != nil {
		NotFound(c, "Sukuk series not found")
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
		InternalServerError(c, "Failed to update sukuk series")
		return
	}

	// Load with company data
	db.Preload("Company").First(&sukuk, sukuk.ID)

	SendSuccess(c, http.StatusOK, sukuk, "Sukuk series updated")
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
// @Success 200 {object} UploadResponse "Prospectus uploaded"
// @Failure 400 {object} APIResponse "Bad request"
// @Failure 404 {object} APIResponse "Sukuk series not found"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /admin/sukuks/{id}/upload-prospectus [post]
func UploadProspectus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid sukuk series ID")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		BadRequest(c, "No file provided")
		return
	}

	// Check if sukuk series exists
	var sukuk models.Sukuk
	db := database.GetDB()
	if err := db.First(&sukuk, uint(id)).Error; err != nil {
		NotFound(c, "Sukuk series not found")
		return
	}

	// Configure file upload
	config := utils.DefaultPDFConfig("./uploads/prospectus")

	// Save file with validation
	filename, url, err := utils.SaveFile(file, config, strconv.FormatUint(id, 10))
	if err != nil {
		BadRequestWithDetails(c, "File upload failed", err.Error())
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
		InternalServerError(c, "Failed to update prospectus URL")
		return
	}

	uploadResponse := UploadResponse{
		Success:  true,
		Message:  "Prospectus uploaded successfully",
		Filename: filename,
		URL:      sukuk.Prospectus,
	}

	c.JSON(http.StatusOK, uploadResponse)
}