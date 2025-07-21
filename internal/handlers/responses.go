package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Standard API Response Models

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError represents an error response
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// PaginatedResponse represents a paginated response
type PaginatedResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Meta    *Pagination `json:"meta"`
}

// Pagination represents pagination metadata
type Pagination struct {
	Total       int `json:"total"`
	Count       int `json:"count"`
	Page        int `json:"page"`
	PerPage     int `json:"per_page"`
	TotalPages  int `json:"total_pages"`
	HasNext     bool `json:"has_next"`
	HasPrevious bool `json:"has_previous"`
}

// Specific Response Types

// CompanyResponse represents a company response
type CompanyResponse struct {
	Success bool            `json:"success"`
	Data    interface{}     `json:"data"` // models.Company or []models.Company
	Meta    *Pagination     `json:"meta,omitempty"`
}

// SukukResponse represents a sukuk response
type SukukResponse struct {
	Success bool            `json:"success"`
	Data    interface{}     `json:"data"` // models.Sukuk or []models.Sukuk
	Meta    *Pagination     `json:"meta,omitempty"`
}

// InvestmentResponse represents an investment response
type InvestmentResponse struct {
	Success bool            `json:"success"`
	Data    interface{}     `json:"data"` // models.Investment or []models.Investment
	Meta    *Pagination     `json:"meta,omitempty"`
}

// YieldResponse represents a yield response
type YieldResponse struct {
	Success bool            `json:"success"`
	Data    interface{}     `json:"data"` // models.Yield or []models.Yield
	Meta    *Pagination     `json:"meta,omitempty"`
}

// RedemptionResponse represents a redemption response
type RedemptionResponse struct {
	Success bool            `json:"success"`
	Data    interface{}     `json:"data"` // models.Redemption or []models.Redemption
	Meta    *Pagination     `json:"meta,omitempty"`
}

// SystemResponse represents a system response
type SystemResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// UploadResponse represents a file upload response
type UploadResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	Filename string `json:"filename,omitempty"`
	URL      string `json:"url,omitempty"`
}

// Centralized Response Functions

// SendSuccess sends a successful response
func SendSuccess(c *gin.Context, code int, data interface{}, message string) {
	response := APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	c.JSON(code, response)
}

// SendError sends an error response
func SendError(c *gin.Context, code int, message string, details string) {
	response := APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	}
	c.JSON(code, response)
}

// SendPaginatedResponse sends a paginated response
func SendPaginatedResponse(c *gin.Context, data interface{}, pagination *Pagination) {
	response := PaginatedResponse{
		Success: true,
		Data:    data,
		Meta:    pagination,
	}
	c.JSON(http.StatusOK, response)
}

// Common Error Handlers

// BadRequest sends a 400 error
func BadRequest(c *gin.Context, message string) {
	SendError(c, http.StatusBadRequest, message, "")
}

// BadRequestWithDetails sends a 400 error with details
func BadRequestWithDetails(c *gin.Context, message string, details string) {
	SendError(c, http.StatusBadRequest, message, details)
}

// NotFound sends a 404 error
func NotFound(c *gin.Context, message string) {
	SendError(c, http.StatusNotFound, message, "")
}

// InternalServerError sends a 500 error
func InternalServerError(c *gin.Context, message string) {
	SendError(c, http.StatusInternalServerError, message, "")
}

// InternalServerErrorWithDetails sends a 500 error with details
func InternalServerErrorWithDetails(c *gin.Context, message string, details string) {
	SendError(c, http.StatusInternalServerError, message, details)
}

// Unauthorized sends a 401 error
func Unauthorized(c *gin.Context, message string) {
	SendError(c, http.StatusUnauthorized, message, "")
}

// Forbidden sends a 403 error
func Forbidden(c *gin.Context, message string) {
	SendError(c, http.StatusForbidden, message, "")
}

// ValidationError sends a 422 error for validation failures
func ValidationError(c *gin.Context, details string) {
	SendError(c, http.StatusUnprocessableEntity, "Validation failed", details)
}