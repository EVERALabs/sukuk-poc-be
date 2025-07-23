package handlers

import (
	"net/http"
	"strconv"

	"sukuk-be/internal/database"
	"sukuk-be/internal/logger"
	"sukuk-be/internal/models"

	"github.com/gin-gonic/gin"
)

// ListSukukMetadata returns all sukuk metadata with optional filtering
// @Summary List sukuk metadata
// @Description Get all sukuk metadata, optionally filter by ready status
// @Tags sukuk-metadata
// @Accept json
// @Produce json
// @Param ready query boolean false "Filter by metadata_ready status"
// @Success 200 {array} models.SukukMetadataResponse
// @Failure 500 {object} map[string]string
// @Router /sukuk-metadata [get]
func ListSukukMetadata(c *gin.Context) {
	var sukukMetadata []models.SukukMetadata
	
	// Check if filtering by ready status
	readyFilter := c.Query("ready")
	query := database.GetDB()
	
	if readyFilter == "true" {
		query = query.Where("metadata_ready = ?", true)
	} else if readyFilter == "false" {
		query = query.Where("metadata_ready = ?", false)
	}
	// If no filter, return all sukuk metadata
	
	result := query.Find(&sukukMetadata)
	if result.Error != nil {
		logger.WithError(result.Error).Error("Failed to fetch sukuk metadata")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch sukuk metadata",
		})
		return
	}

	// Convert to response format
	responses := make([]*models.SukukMetadataResponse, len(sukukMetadata))
	for i, sukuk := range sukukMetadata {
		responses[i] = sukuk.ToResponse()
	}

	c.JSON(http.StatusOK, responses)
}

// CreateSukukMetadata creates new sukuk metadata
// @Summary Create sukuk metadata
// @Description Create new sukuk with onchain and offchain metadata
// @Tags sukuk-metadata
// @Accept json
// @Produce json
// @Param sukuk body models.SukukMetadataCreateRequest true "Sukuk metadata"
// @Success 201 {object} models.SukukMetadataResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /sukuk-metadata [post]
func CreateSukukMetadata(c *gin.Context) {
	var req models.SukukMetadataCreateRequest
	
	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithError(err).Error("Invalid request payload")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Create sukuk metadata model
	sukukMetadata := models.SukukMetadata{
		// Onchain Data
		ContractAddress: req.ContractAddress,
		TokenID:         req.TokenID,
		OwnerAddress:    req.OwnerAddress,
		TransactionHash: req.TransactionHash,
		BlockNumber:     req.BlockNumber,
		
		// Basic Info
		SukukCode:      req.SukukCode,
		SukukTitle:     req.SukukTitle,
		SukukDeskripsi: req.SukukDeskripsi,
		Status:         req.Status,
		LogoURL:        req.LogoURL,
		
		// Main Features
		Tenor:       req.Tenor,
		ImbalHasil:  req.ImbalHasil,
		
		// Ketentuan
		PeriodePembelian:  req.PeriodePembelian,
		JatuhTempo:        req.JatuhTempo,
		KuotaNasional:     req.KuotaNasional,
		PenerimaanKupon:   req.PenerimaanKupon,
		MinimumPembelian:  req.MinimumPembelian,
		TanggalBayarKupon: req.TanggalBayarKupon,
		MaksimumPembelian: req.MaksimumPembelian,
		KuponPertama:      req.KuponPertama,
		TipeKupon:         req.TipeKupon,
		
		// Default to not ready
		MetadataReady: false,
	}

	// Create in database
	if err := database.GetDB().Create(&sukukMetadata).Error; err != nil {
		logger.WithError(err).Error("Failed to create sukuk metadata")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create sukuk metadata",
		})
		return
	}

	logger.WithFields(map[string]interface{}{
		"sukuk_code": sukukMetadata.SukukCode,
		"id":         sukukMetadata.ID,
	}).Info("Sukuk metadata created successfully")

	c.JSON(http.StatusCreated, sukukMetadata.ToResponse())
}

// MarkSukukMetadataReady marks sukuk metadata as ready
// @Summary Mark sukuk metadata as ready
// @Description Update sukuk metadata_ready flag to true
// @Tags sukuk-metadata
// @Accept json
// @Produce json
// @Param id path string true "Sukuk metadata ID"
// @Success 200 {object} models.SukukMetadataResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /sukuk-metadata/{id}/ready [put]
func MarkSukukMetadataReady(c *gin.Context) {
	// Get ID from path
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID format",
		})
		return
	}

	// Find sukuk metadata
	var sukukMetadata models.SukukMetadata
	result := database.GetDB().First(&sukukMetadata, "id = ?", uint(id))
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Sukuk metadata not found",
		})
		return
	}

	// Update metadata_ready flag
	result = database.GetDB().Model(&sukukMetadata).Update("metadata_ready", true)
	if result.Error != nil {
		logger.WithError(result.Error).Error("Failed to update sukuk metadata")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update sukuk metadata",
		})
		return
	}

	// Reload the updated model
	database.GetDB().First(&sukukMetadata, "id = ?", uint(id))

	logger.WithFields(map[string]interface{}{
		"sukuk_code": sukukMetadata.SukukCode,
		"id":         sukukMetadata.ID,
	}).Info("Sukuk metadata marked as ready")

	c.JSON(http.StatusOK, sukukMetadata.ToResponse())
}

// UpdateSukukMetadata updates sukuk metadata with offchain data
// @Summary Update sukuk metadata
// @Description Update existing sukuk metadata with offchain data
// @Tags sukuk-metadata
// @Accept json
// @Produce json
// @Param id path string true "Sukuk metadata ID"
// @Param sukuk body models.SukukMetadataUpdateRequest true "Sukuk metadata updates"
// @Success 200 {object} models.SukukMetadataResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /sukuk-metadata/{id} [put]
func UpdateSukukMetadata(c *gin.Context) {
	// Get ID from path
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid ID format",
		})
		return
	}

	// Bind and validate request
	var req models.SukukMetadataUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithError(err).Error("Invalid request payload")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request payload",
			"details": err.Error(),
		})
		return
	}

	// Find sukuk metadata
	var sukukMetadata models.SukukMetadata
	result := database.GetDB().First(&sukukMetadata, "id = ?", uint(id))
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Sukuk metadata not found",
		})
		return
	}

	// Update fields if provided
	if req.SukukTitle != nil {
		sukukMetadata.SukukTitle = *req.SukukTitle
	}
	if req.SukukDeskripsi != nil {
		sukukMetadata.SukukDeskripsi = *req.SukukDeskripsi
	}
	if req.Status != nil {
		sukukMetadata.Status = *req.Status
	}
	if req.LogoURL != nil {
		sukukMetadata.LogoURL = *req.LogoURL
	}
	if req.Tenor != nil {
		sukukMetadata.Tenor = *req.Tenor
	}
	if req.ImbalHasil != nil {  
		sukukMetadata.ImbalHasil = *req.ImbalHasil
	}
	if req.PeriodePembelian != nil {
		sukukMetadata.PeriodePembelian = *req.PeriodePembelian
	}
	if req.JatuhTempo != nil {
		sukukMetadata.JatuhTempo = *req.JatuhTempo
	}
	if req.KuotaNasional != nil {
		sukukMetadata.KuotaNasional = *req.KuotaNasional
	}
	if req.PenerimaanKupon != nil {
		sukukMetadata.PenerimaanKupon = *req.PenerimaanKupon
	}
	if req.MinimumPembelian != nil {
		sukukMetadata.MinimumPembelian = *req.MinimumPembelian
	}
	if req.TanggalBayarKupon != nil {
		sukukMetadata.TanggalBayarKupon = *req.TanggalBayarKupon
	}
	if req.MaksimumPembelian != nil {
		sukukMetadata.MaksimumPembelian = *req.MaksimumPembelian
	}
	if req.KuponPertama != nil {
		sukukMetadata.KuponPertama = *req.KuponPertama
	}
	if req.TipeKupon != nil {
		sukukMetadata.TipeKupon = *req.TipeKupon
	}

	// Save updates
	if err := database.GetDB().Save(&sukukMetadata).Error; err != nil {
		logger.WithError(err).Error("Failed to update sukuk metadata")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update sukuk metadata",
		})
		return
	}

	logger.WithFields(map[string]interface{}{
		"sukuk_code": sukukMetadata.SukukCode,
		"id":         sukukMetadata.ID,
	}).Info("Sukuk metadata updated successfully")

	c.JSON(http.StatusOK, sukukMetadata.ToResponse())
}