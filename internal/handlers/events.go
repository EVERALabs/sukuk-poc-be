package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kadzu/sukuk-poc-be/internal/database"
	"github.com/kadzu/sukuk-poc-be/internal/models"
	"gorm.io/gorm"
)

// WebhookEventRequest represents the request body for processing blockchain events
type WebhookEventRequest struct {
	EventName       string                 `json:"event_name" binding:"required" swaggertype:"string" example:"Investment"`
	BlockNumber     int64                  `json:"block_number" binding:"required" swaggertype:"integer" example:"12345678"`
	TxHash          string                 `json:"tx_hash" binding:"required" swaggertype:"string" example:"0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890"`
	ContractAddress string                 `json:"contract_address" binding:"required" swaggertype:"string" example:"0xabcdefabcdefabcdefabcdefabcdefabcdefabcd"`
	Data            map[string]interface{} `json:"data" binding:"required" swaggertype:"object" example:"{\"investor\":\"0x123...\",\"amount\":\"1000\"}"`
	ChainID         int64                  `json:"chain_id" swaggertype:"integer" example:"84532"`
}

// ProcessEventWebhook processes blockchain events from indexer (WEBHOOK)
// @Summary Process blockchain event webhook
// @Description Process blockchain events from external indexer (admin only)
// @Tags Events
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body WebhookEventRequest true "Event data from blockchain indexer"
// @Success 200 {object} EventWebhookResponse "Event processed successfully"
// @Failure 400 {object} ErrorResponse "Bad request"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /admin/events/webhook [post]
func ProcessEventWebhook(c *gin.Context) {
	var req WebhookEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default chain ID if not provided
	if req.ChainID == 0 {
		req.ChainID = 84532 // Base Testnet
	}

	// Create event record
	event := models.Event{
		EventName:       req.EventName,
		BlockNumber:     req.BlockNumber,
		TxHash:          req.TxHash,
		ContractAddress: req.ContractAddress,
		Data:            models.JSON(req.Data),
		ChainID:         req.ChainID,
		Processed:       false,
	}

	db := database.GetDB()
	if err := db.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to store event",
		})
		return
	}

	// Process event based on type
	switch req.EventName {
	case "Investment":
		processInvestmentEvent(db, &event, req.Data)
	case "YieldClaim":
		processYieldClaimEvent(db, &event, req.Data)
	case "Redemption":
		processRedemptionEvent(db, &event, req.Data)
	case "SukukDeployment":
		processSukukDeploymentEvent(db, &event, req.Data)
	default:
		// Mark as processed for unknown events
		event.MarkAsProcessed(db)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Event processed successfully",
		"event_id": event.ID,
	})
}

// GetEventByTxHash returns event details by transaction hash (READ-ONLY)
// @Summary Get events by transaction hash
// @Description Get all blockchain events associated with a specific transaction hash
// @Tags Events
// @Accept json
// @Produce json
// @Param txHash path string true "Transaction hash"
// @Success 200 {object} EventListResponse "List of events for transaction"
// @Failure 400 {object} ErrorResponse "Invalid transaction hash"
// @Failure 404 {object} ErrorResponse "No events found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /events/{txHash} [get]
func GetEventByTxHash(c *gin.Context) {
	txHash := c.Param("txHash")
	if txHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid transaction hash",
		})
		return
	}

	var events []models.Event
	db := database.GetDB()
	if err := db.Where("tx_hash = ?", txHash).Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch events",
		})
		return
	}

	if len(events) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "No events found for this transaction",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": events,
		"count": len(events),
	})
}

// Helper functions to process different event types
func processInvestmentEvent(db *gorm.DB, event *models.Event, data map[string]interface{}) {
	// Extract investment data from event
	investorAddress, _ := data["investor"].(string)
	sukukSeriesID, _ := data["sukuk_series_id"].(float64)
	amount, _ := data["amount"].(string)
	tokenAmount, _ := data["token_amount"].(string)

	investment := models.Investment{
		SukukSeriesID:   uint(sukukSeriesID),
		InvestorAddress: investorAddress,
		Amount:          amount,
		TokenAmount:     tokenAmount,
		Status:          models.InvestmentStatusActive,
		TransactionHash: event.TxHash,
		BlockNumber:     uint64(event.BlockNumber),
		InvestmentDate:  event.CreatedAt,
	}

	if err := db.Create(&investment).Error; err != nil {
		event.MarkAsError(db, "Failed to create investment record: "+err.Error())
		return
	}

	event.MarkAsProcessed(db)
}

func processYieldClaimEvent(db *gorm.DB, event *models.Event, data map[string]interface{}) {
	// Extract yield claim data from event
	yieldClaimID, _ := data["yield_claim_id"].(float64)

	// Update existing yield claim as claimed
	var yieldClaim models.YieldClaim
	if err := db.First(&yieldClaim, uint(yieldClaimID)).Error; err != nil {
		event.MarkAsError(db, "Yield claim not found: "+err.Error())
		return
	}

	now := event.CreatedAt
	yieldClaim.Status = models.YieldClaimStatusClaimed
	yieldClaim.ClaimedAt = &now
	yieldClaim.TransactionHash = event.TxHash
	yieldClaim.BlockNumber = uint64(event.BlockNumber)

	if err := db.Save(&yieldClaim).Error; err != nil {
		event.MarkAsError(db, "Failed to update yield claim: "+err.Error())
		return
	}

	event.MarkAsProcessed(db)
}

func processRedemptionEvent(db *gorm.DB, event *models.Event, data map[string]interface{}) {
	// Extract redemption data from event
	redemptionID, _ := data["redemption_id"].(float64)

	// Update existing redemption as completed
	var redemption models.Redemption
	if err := db.First(&redemption, uint(redemptionID)).Error; err != nil {
		event.MarkAsError(db, "Redemption not found: "+err.Error())
		return
	}

	now := event.CreatedAt
	redemption.Status = models.RedemptionStatusCompleted
	redemption.CompletedAt = &now
	redemption.TransactionHash = event.TxHash
	redemption.BlockNumber = uint64(event.BlockNumber)

	if err := db.Save(&redemption).Error; err != nil {
		event.MarkAsError(db, "Failed to update redemption: "+err.Error())
		return
	}

	event.MarkAsProcessed(db)
}

func processSukukDeploymentEvent(db *gorm.DB, event *models.Event, data map[string]interface{}) {
	// Extract deployment data from event
	contractAddress, _ := data["contract_address"].(string)
	seriesName, _ := data["series_name"].(string)

	// Find sukuk series by name and update contract address
	var sukukSeries models.SukukSeries
	if err := db.Where("name = ? AND token_address = ''", seriesName).First(&sukukSeries).Error; err != nil {
		event.MarkAsError(db, "Sukuk series not found for deployment: "+err.Error())
		return
	}

	sukukSeries.TokenAddress = contractAddress
	sukukSeries.Status = models.SukukStatusActive

	if err := db.Save(&sukukSeries).Error; err != nil {
		event.MarkAsError(db, "Failed to update sukuk series: "+err.Error())
		return
	}

	event.MarkAsProcessed(db)
}