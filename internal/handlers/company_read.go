package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/kadzu/sukuk-poc-be/internal/database"
	"github.com/kadzu/sukuk-poc-be/internal/models"
)

// ListCompanies godoc
// @Summary List all companies
// @Description Get a list of all active companies
// @Tags companies
// @Accept json
// @Produce json
// @Param sector query string false "Filter by industry sector"
// @Success 200 {object} CompanyListResponse "List of companies"
// @Failure 500 {object} ErrorResponse "Internal server error"
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
		"data": companies,
		"count": len(companies),
		"meta": gin.H{
			"total": len(companies),
			"page": 1,
		},
	})
}

// GetCompany godoc
// @Summary Get company details
// @Description Get details of a specific company including its Sukuk series
// @Tags companies
// @Accept json
// @Produce json
// @Param id path int true "Company ID"
// @Success 200 {object} CompanyResponse "Company details"
// @Failure 400 {object} ErrorResponse "Invalid company ID"
// @Failure 404 {object} ErrorResponse "Company not found"
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
	if err := db.Preload("SukukSeries").First(&company, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Company not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": company,
	})
}

// GetCompanySukuks godoc
// @Summary Get company's Sukuk series
// @Description Get all Sukuk series for a specific company
// @Tags companies
// @Accept json
// @Produce json
// @Param id path int true "Company ID"
// @Success 200 {object} CompanySukuksResponse "List of company's Sukuk series"
// @Failure 400 {object} ErrorResponse "Invalid company ID"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /companies/{id}/sukuks [get]
func GetCompanySukuks(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid company ID",
		})
		return
	}

	var sukukSeries []models.SukukSeries
	db := database.GetDB()
	if err := db.Preload("Company").Where("company_id = ?", uint(id)).Find(&sukukSeries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch company sukuks",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": sukukSeries,
		"count": len(sukukSeries),
	})
}