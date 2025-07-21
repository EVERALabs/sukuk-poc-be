package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"sukuk-be/internal/database"
	"sukuk-be/internal/models"
	"sukuk-be/internal/utils"

	"github.com/gin-gonic/gin"
)

// Request/Response Types
type CreateCompanyRequest struct {
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description"`
	Website       string `json:"website"`
	Industry      string `json:"industry"`
	Email         string `json:"email" binding:"required,email"`
	WalletAddress string `json:"wallet_address" binding:"required"`
}

type UpdateCompanyRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Website     string `json:"website,omitempty"`
	Industry    string `json:"industry,omitempty"`
	Email       string `json:"email,omitempty"`
	IsActive    *bool  `json:"is_active,omitempty"`
}

// Public Company APIs

// ListCompanies returns all active companies
// @Summary List all companies
// @Description Get a list of all active companies
// @Tags Companies
// @Accept json
// @Produce json
// @Param sector query string false "Filter by industry sector"
// @Success 200 {object} CompanyResponse "List of companies"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /companies [get]
func ListCompanies(c *gin.Context) {
	var companies []models.Company

	db := database.GetDB()
	query := db.Where("is_active = ?", true)

	// Filter by sector if provided
	if sector := c.Query("sector"); sector != "" {
		query = query.Where("industry = ?", sector)
	}

	if err := query.Find(&companies).Error; err != nil {
		InternalServerErrorWithDetails(c, "Failed to fetch companies", err.Error())
		return
	}

	pagination := &Pagination{
		Total:   len(companies),
		Count:   len(companies),
		Page:    1,
		PerPage: len(companies),
		TotalPages: 1,
		HasNext: false,
		HasPrevious: false,
	}

	SendPaginatedResponse(c, companies, pagination)
}

// GetCompany returns company details by ID
// @Summary Get company details
// @Description Get details of a specific company including its Sukuk series
// @Tags Companies
// @Accept json
// @Produce json
// @Param id path int true "Company ID"
// @Success 200 {object} CompanyResponse "Company details"
// @Failure 400 {object} APIResponse "Invalid company ID"
// @Failure 404 {object} APIResponse "Company not found"
// @Router /companies/{id} [get]
func GetCompany(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid company ID")
		return
	}

	var company models.Company
	db := database.GetDB()
	if err := db.Preload("Sukuks").First(&company, uint(id)).Error; err != nil {
		NotFound(c, "Company not found")
		return
	}

	SendSuccess(c, http.StatusOK, company, "Company retrieved successfully")
}

// GetCompanySukuks returns all Sukuk series for a company
// @Summary Get company's Sukuk series
// @Description Get all Sukuk series for a specific company
// @Tags Companies
// @Accept json
// @Produce json
// @Param id path int true "Company ID"
// @Success 200 {object} SukukResponse "List of company's Sukuk series"
// @Failure 400 {object} APIResponse "Invalid company ID"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /companies/{id}/sukuks [get]
func GetCompanySukuks(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid company ID")
		return
	}

	var sukuks []models.Sukuk
	db := database.GetDB()
	if err := db.Preload("Company").Where("company_id = ?", uint(id)).Find(&sukuks).Error; err != nil {
		InternalServerErrorWithDetails(c, "Failed to fetch company sukuks", err.Error())
		return
	}

	pagination := &Pagination{
		Total:       len(sukuks),
		Count:       len(sukuks),
		Page:        1,
		PerPage:     len(sukuks),
		TotalPages:  1,
		HasNext:     false,
		HasPrevious: false,
	}

	SendPaginatedResponse(c, sukuks, pagination)
}

// Admin Company Management APIs

// CreateCompany creates a new company
// @Summary Create a new company
// @Description Create a new partner company (Admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Param company body CreateCompanyRequest true "Company data"
// @Success 201 {object} CompanyResponse "Company created successfully"
// @Failure 400 {object} APIResponse "Invalid request data"
// @Failure 500 {object} APIResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /admin/companies [post]
func CreateCompany(c *gin.Context) {
	var req CreateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequestWithDetails(c, "Invalid request data", err.Error())
		return
	}

	// Validate wallet address format
	if !utils.IsValidEthereumAddress(req.WalletAddress) {
		BadRequest(c, "Invalid Ethereum address format")
		return
	}

	company := models.Company{
		Name:          req.Name,
		Description:   req.Description,
		Website:       req.Website,
		Industry:      req.Industry,
		Email:         req.Email,
		WalletAddress: strings.ToLower(req.WalletAddress),
		IsActive:      true,
	}

	db := database.GetDB()
	if err := db.Create(&company).Error; err != nil {
		InternalServerErrorWithDetails(c, "Failed to create company", err.Error())
		return
	}

	SendSuccess(c, http.StatusCreated, company, "Company created successfully")
}

// UpdateCompany updates a company by ID
// @Summary Update company
// @Description Update company details (Admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path int true "Company ID"
// @Param company body UpdateCompanyRequest true "Updated company data"
// @Success 200 {object} CompanyResponse "Company updated successfully"
// @Failure 400 {object} APIResponse "Invalid request"
// @Failure 404 {object} APIResponse "Company not found"
// @Failure 500 {object} APIResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /admin/companies/{id} [put]
func UpdateCompany(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid company ID")
		return
	}

	var req UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequestWithDetails(c, "Invalid request data", err.Error())
		return
	}

	db := database.GetDB()
	var company models.Company
	if err := db.First(&company, uint(id)).Error; err != nil {
		NotFound(c, "Company not found")
		return
	}

	// Update fields if provided
	if req.Name != "" {
		company.Name = req.Name
	}
	if req.Description != "" {
		company.Description = req.Description
	}
	if req.Website != "" {
		company.Website = req.Website
	}
	if req.Industry != "" {
		company.Industry = req.Industry
	}
	if req.Email != "" {
		company.Email = req.Email
	}
	if req.IsActive != nil {
		company.IsActive = *req.IsActive
	}

	if err := db.Save(&company).Error; err != nil {
		InternalServerErrorWithDetails(c, "Failed to update company", err.Error())
		return
	}

	SendSuccess(c, http.StatusOK, company, "Company updated successfully")
}

// UploadCompanyLogo uploads a logo for a company
// @Summary Upload company logo
// @Description Upload logo image for a company (Admin only)
// @Tags Admin
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Company ID"
// @Param file formData file true "Logo image file (PNG, JPG, JPEG, max 5MB)"
// @Success 200 {object} UploadResponse "Logo uploaded successfully"
// @Failure 400 {object} APIResponse "Invalid file or company ID"
// @Failure 404 {object} APIResponse "Company not found"
// @Failure 500 {object} APIResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /admin/companies/{id}/upload-logo [post]
func UploadCompanyLogo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid company ID")
		return
	}

	// Verify company exists
	db := database.GetDB()
	var company models.Company
	if err := db.First(&company, uint(id)).Error; err != nil {
		NotFound(c, "Company not found")
		return
	}

	// Handle file upload
	file, err := c.FormFile("file")
	if err != nil {
		BadRequest(c, "No file provided")
		return
	}

	// Validate file type
	allowedTypes := []string{".png", ".jpg", ".jpeg"}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	validType := false
	for _, allowedType := range allowedTypes {
		if ext == allowedType {
			validType = true
			break
		}
	}

	if !validType {
		BadRequest(c, "File type not allowed. Only PNG, JPG, and JPEG files are supported")
		return
	}

	// Validate file size (5MB max)
	const maxSize = 5 << 20 // 5MB
	if file.Size > maxSize {
		BadRequest(c, "File size too large. Maximum allowed size is 5MB")
		return
	}

	// Create uploads directory if it doesn't exist
	uploadDir := "uploads/logos"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		InternalServerErrorWithDetails(c, "Failed to create upload directory", err.Error())
		return
	}

	// Generate unique filename
	filename := "company_" + strconv.FormatUint(id, 10) + "_logo" + ext
	filePath := filepath.Join(uploadDir, filename)
	urlPath := "/" + filePath

	// Save file
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		InternalServerErrorWithDetails(c, "Failed to save file", err.Error())
		return
	}

	// Update company logo path in database
	company.Logo = urlPath
	if err := db.Save(&company).Error; err != nil {
		InternalServerErrorWithDetails(c, "Failed to update company record", err.Error())
		return
	}

	response := UploadResponse{
		Success:  true,
		Message:  "Company logo uploaded successfully",
		Filename: filename,
		URL:      urlPath,
	}
	c.JSON(http.StatusOK, response)
}