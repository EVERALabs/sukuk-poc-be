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
// @Success 200 {object} map[string]interface{} "List of companies"
// @Failure 500 {object} map[string]interface{} "Internal server error"
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch companies",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  companies,
		"count": len(companies),
		"meta": gin.H{
			"total": len(companies),
			"page":  1,
		},
	})
}

// GetCompany returns company details by ID
// @Summary Get company details
// @Description Get details of a specific company including its Sukuk series
// @Tags Companies
// @Accept json
// @Produce json
// @Param id path int true "Company ID"
// @Success 200 {object} map[string]interface{} "Company details"
// @Failure 400 {object} map[string]interface{} "Invalid company ID"
// @Failure 404 {object} map[string]interface{} "Company not found"
// @Router /companies/{id} [get]
func GetCompany(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid company ID",
		})
		return
	}

	var company models.Company
	db := database.GetDB()
	if err := db.Preload("Sukuks").First(&company, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Company not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": company,
	})
}

// GetCompanySukuks returns all Sukuk series for a company
// @Summary Get company's Sukuk series
// @Description Get all Sukuk series for a specific company
// @Tags Companies
// @Accept json
// @Produce json
// @Param id path int true "Company ID"
// @Success 200 {object} map[string]interface{} "List of company's Sukuk series"
// @Failure 400 {object} map[string]interface{} "Invalid company ID"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /companies/{id}/sukuks [get]
func GetCompanySukuks(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid company ID",
		})
		return
	}

	var sukuks []models.Sukuk
	db := database.GetDB()
	if err := db.Preload("Company").Where("company_id = ?", uint(id)).Find(&sukuks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch company sukuks",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  sukuks,
		"count": len(sukuks),
	})
}

// Admin Company Management APIs

// CreateCompany creates a new company
// @Summary Create a new company
// @Description Create a new partner company (Admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Param company body CreateCompanyRequest true "Company data"
// @Success 201 {object} map[string]interface{} "Company created successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request data"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security ApiKeyAuth
// @Router /admin/companies [post]
func CreateCompany(c *gin.Context) {
	var req CreateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate wallet address format
	if !utils.IsValidEthereumAddress(req.WalletAddress) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid Ethereum address format",
		})
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create company",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Company created successfully",
		"data":    company,
	})
}

// UpdateCompany updates a company by ID
// @Summary Update company
// @Description Update company details (Admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Param id path int true "Company ID"
// @Param company body UpdateCompanyRequest true "Updated company data"
// @Success 200 {object} map[string]interface{} "Company updated successfully"
// @Failure 400 {object} map[string]interface{} "Invalid request"
// @Failure 404 {object} map[string]interface{} "Company not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security ApiKeyAuth
// @Router /admin/companies/{id} [put]
func UpdateCompany(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid company ID",
		})
		return
	}

	var req UpdateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	var company models.Company
	if err := db.First(&company, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Company not found",
		})
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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update company",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Company updated successfully",
		"data":    company,
	})
}

// UploadCompanyLogo uploads a logo for a company
// @Summary Upload company logo
// @Description Upload logo image for a company (Admin only)
// @Tags Admin
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "Company ID"
// @Param file formData file true "Logo image file (PNG, JPG, JPEG, max 5MB)"
// @Success 200 {object} map[string]interface{} "Logo uploaded successfully"
// @Failure 400 {object} map[string]interface{} "Invalid file or company ID"
// @Failure 404 {object} map[string]interface{} "Company not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Security ApiKeyAuth
// @Router /admin/companies/{id}/upload-logo [post]
func UploadCompanyLogo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid company ID",
		})
		return
	}

	// Verify company exists
	db := database.GetDB()
	var company models.Company
	if err := db.First(&company, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Company not found",
		})
		return
	}

	// Handle file upload
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No file provided",
		})
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file type not allowed",
		})
		return
	}

	// Validate file size (5MB max)
	const maxSize = 5 << 20 // 5MB
	if file.Size > maxSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "file size too large",
		})
		return
	}

	// Create uploads directory if it doesn't exist
	uploadDir := "uploads/logos"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create upload directory",
		})
		return
	}

	// Generate unique filename
	filename := "company_" + strconv.FormatUint(id, 10) + "_logo" + ext
	filePath := filepath.Join(uploadDir, filename)
	urlPath := "/" + filePath

	// Save file
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save file",
		})
		return
	}

	// Update company logo path in database
	company.Logo = urlPath
	if err := db.Save(&company).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update company record",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Company logo uploaded successfully",
		"filename": filename,
		"url":      urlPath,
	})
}