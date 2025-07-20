package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kadzu/sukuk-poc-be/internal/database"
	"github.com/kadzu/sukuk-poc-be/internal/models"
	"github.com/kadzu/sukuk-poc-be/internal/utils"
)

type CreateCompanyRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Website     string `json:"website"`
	Industry    string `json:"industry"`
	Email       string `json:"email" binding:"required,email"`
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

// CreateCompany creates a new partner company
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