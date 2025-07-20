package handlers

import (
	"net/http"
	"strconv"

	"sukuk-be/internal/database"
	"sukuk-be/internal/models"
	"sukuk-be/internal/utils"

	"github.com/gin-gonic/gin"
)

// CreateCompanyRequest represents the request body for creating a new company
type CreateCompanyRequest struct {
	Name          string `json:"name" binding:"required" swaggertype:"string" example:"PT Sukuk Indonesia"`
	Description   string `json:"description" swaggertype:"string" example:"Leading Indonesian sukuk issuer"`
	Website       string `json:"website" swaggertype:"string" example:"https://sukukindonesia.com"`
	Industry      string `json:"industry" swaggertype:"string" example:"Financial Services"`
	Email         string `json:"email" binding:"required,email" swaggertype:"string" example:"contact@sukukindonesia.com"`
	WalletAddress string `json:"wallet_address" binding:"required" swaggertype:"string" example:"0x1234567890123456789012345678901234567890"`
}

// UpdateCompanyRequest represents the request body for updating a company
type UpdateCompanyRequest struct {
	Name        string `json:"name,omitempty" swaggertype:"string" example:"PT Sukuk Indonesia Updated"`
	Description string `json:"description,omitempty" swaggertype:"string" example:"Updated description"`
	Website     string `json:"website,omitempty" swaggertype:"string" example:"https://newsukukindonesia.com"`
	Industry    string `json:"industry,omitempty" swaggertype:"string" example:"Financial Technology"`
	Email       string `json:"email,omitempty" swaggertype:"string" example:"newcontact@sukukindonesia.com"`
	IsActive    *bool  `json:"is_active,omitempty" swaggertype:"boolean" example:"true"`
}

// CreateCompany godoc
// @Summary Create a new company
// @Description Create a new partner company (Admin only)
// @Tags admin
// @Accept json
// @Produce json
// @Param company body CreateCompanyRequest true "Company data"
// @Success 201 {object} APIResponse "Company created successfully"
// @Failure 400 {object} ErrorResponse "Invalid request data"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Security ApiKeyAuth
// @Router /admin/companies [post]
func CreateCompany(c *gin.Context) {
	var req CreateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	company := models.Company{
		Name:          req.Name,
		Description:   req.Description,
		Website:       req.Website,
		Industry:      req.Industry,
		Email:         req.Email,
		WalletAddress: req.WalletAddress,
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

// UpdateCompany updates existing company information
// @Summary Update company
// @Description Update existing company information (admin only)
// @Tags Company Management
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Company ID"
// @Param request body UpdateCompanyRequest true "Update data"
// @Success 200 {object} APIResponse "Company updated"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 404 {object} ErrorResponse "Company not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
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

	var company models.Company
	db := database.GetDB()
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

// UploadCompanyLogo handles company logo file upload
// @Summary Upload company logo
// @Description Upload logo image file for a company (admin only)
// @Tags Company Management
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Company ID"
// @Param file formData file true "Logo image file (jpg, png, gif)"
// @Success 200 {object} FileUploadResponse "Logo uploaded"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 404 {object} ErrorResponse "Company not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /admin/companies/{id}/upload-logo [post]
func UploadCompanyLogo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid company ID",
		})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file provided"})
		return
	}

	// Check if company exists
	var company models.Company
	db := database.GetDB()
	if err := db.First(&company, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Company not found",
		})
		return
	}

	// Configure file upload
	config := utils.DefaultImageConfig("./uploads/logos")

	// Save file with validation
	filename, url, err := utils.SaveFile(file, config, strconv.FormatUint(id, 10))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Delete old logo if exists
	if company.Logo != "" {
		utils.DeleteFile(company.Logo)
	}

	// Update company logo URL
	company.Logo = url
	if err := db.Save(&company).Error; err != nil {
		// Clean up uploaded file if database update fails
		utils.DeleteFile(url)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update company logo URL",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Company logo uploaded successfully",
		"filename": filename,
		"url":      company.Logo,
	})
}
