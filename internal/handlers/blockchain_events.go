package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"sukuk-be/internal/database"
	"sukuk-be/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ProcessSukukPurchasedEvent processes a single SukukPurchased event
// @Summary Process a SukukPurchased event
// @Description Process a single SukukPurchased event from the blockchain
// @Tags Blockchain Events
// @Accept json
// @Produce json
// @Param event body models.SukukPurchased true "SukukPurchased event data"
// @Success 200 {object} APIResponse "Event processed successfully"
// @Failure 400 {object} APIResponse "Invalid request"
// @Failure 409 {object} APIResponse "Event already processed"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /events/sukuk-purchased [post]
func ProcessSukukPurchasedEvent(c *gin.Context) {
	var event models.SukukPurchased
	if err := c.ShouldBindJSON(&event); err != nil {
		BadRequestWithDetails(c, "Invalid request body", err.Error())
		return
	}

	db := database.GetDB()
	// Start a transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if event already exists
	existingEvent, err := models.GetSukukPurchaseByTxHashAndLogIndex(tx, event.TxHash, event.LogIndex)
	if err != nil && err != gorm.ErrRecordNotFound {
		tx.Rollback()
		InternalServerErrorWithDetails(c, "Failed to check existing event", err.Error())
		return
	}

	if existingEvent != nil {
		tx.Rollback()
		SendError(c, http.StatusConflict, "Event already processed", "This event has already been processed")
		return
	}

	// Save the event
	if err := tx.Create(&event).Error; err != nil {
		tx.Rollback()
		InternalServerErrorWithDetails(c, "Failed to save event", err.Error())
		return
	}

	// Process the event - create/update investment record
	if err := processSukukPurchase(tx, &event); err != nil {
		tx.Rollback()
		InternalServerErrorWithDetails(c, "Failed to process event", err.Error())
		return
	}

	// Mark event as processed
	if err := event.MarkAsProcessed(tx); err != nil {
		tx.Rollback()
		InternalServerErrorWithDetails(c, "Failed to mark event as processed", err.Error())
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		InternalServerErrorWithDetails(c, "Failed to commit transaction", err.Error())
		return
	}

	SendSuccess(c, http.StatusOK, event, "SukukPurchased event processed successfully")
}

// ProcessRedemptionRequestedEvent processes a single RedemptionRequested event
// @Summary Process a RedemptionRequested event
// @Description Process a single RedemptionRequested event from the blockchain
// @Tags Blockchain Events
// @Accept json
// @Produce json
// @Param event body models.RedemptionRequested true "RedemptionRequested event data"
// @Success 200 {object} APIResponse "Event processed successfully"
// @Failure 400 {object} APIResponse "Invalid request"
// @Failure 409 {object} APIResponse "Event already processed"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /events/redemption-requested [post]
func ProcessRedemptionRequestedEvent(c *gin.Context) {
	var event models.RedemptionRequested
	if err := c.ShouldBindJSON(&event); err != nil {
		BadRequestWithDetails(c, "Invalid request body", err.Error())
		return
	}

	db := database.GetDB()
	// Start a transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if event already exists
	existingEvent, err := models.GetRedemptionRequestByTxHashAndLogIndex(tx, event.TxHash, event.LogIndex)
	if err != nil && err != gorm.ErrRecordNotFound {
		tx.Rollback()
		InternalServerErrorWithDetails(c, "Failed to check existing event", err.Error())
		return
	}

	if existingEvent != nil {
		tx.Rollback()
		SendError(c, http.StatusConflict, "Event already processed", "This event has already been processed")
		return
	}

	// Save the event
	if err := tx.Create(&event).Error; err != nil {
		tx.Rollback()
		InternalServerErrorWithDetails(c, "Failed to save event", err.Error())
		return
	}

	// Process the event - create redemption record
	if err := processRedemptionRequest(tx, &event); err != nil {
		tx.Rollback()
		InternalServerErrorWithDetails(c, "Failed to process event", err.Error())
		return
	}

	// Mark event as processed
	if err := event.MarkAsProcessed(tx); err != nil {
		tx.Rollback()
		InternalServerErrorWithDetails(c, "Failed to mark event as processed", err.Error())
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		InternalServerErrorWithDetails(c, "Failed to commit transaction", err.Error())
		return
	}

	SendSuccess(c, http.StatusOK, event, "RedemptionRequested event processed successfully")
}

// TriggerEventSync manually triggers blockchain event synchronization
// @Summary Trigger manual event sync
// @Description Manually trigger synchronization of blockchain events from the indexer
// @Tags Blockchain Events
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse "Sync completed"
// @Failure 500 {object} APIResponse "Sync failed"
// @Router /events/sync [post]
func TriggerEventSync(c *gin.Context) {
	// Initialize blockchain event sync service
	// eventSync := services.NewBlockchainEventSync(nil)

	// Trigger manual sync
	// if err := eventSync.SyncEvents(c.Request.Context()); err != nil {
	// 	InternalServerErrorWithDetails(c, "Failed to sync events", err.Error())
	// 	return
	// }

	// SendSuccess(c, http.StatusOK, nil, "Blockchain events synchronized successfully")
}

// GetUnprocessedSukukPurchases returns unprocessed sukuk purchase events
// @Summary Get unprocessed SukukPurchased events
// @Description Retrieve unprocessed SukukPurchased events
// @Tags Blockchain Events
// @Accept json
// @Produce json
// @Param limit query integer false "Number of events to retrieve" default(100)
// @Success 200 {object} APIResponse "List of unprocessed events"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /events/sukuk-purchased/unprocessed [get]
func GetUnprocessedSukukPurchases(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 100
	}

	db := database.GetDB()
	events, err := models.GetUnprocessedSukukPurchases(db, limit)
	if err != nil {
		InternalServerErrorWithDetails(c, "Failed to fetch unprocessed events", err.Error())
		return
	}

	SendSuccess(c, http.StatusOK, events, "")
}

// GetUnprocessedRedemptionRequests returns unprocessed redemption request events
// @Summary Get unprocessed RedemptionRequested events
// @Description Retrieve unprocessed RedemptionRequested events
// @Tags Blockchain Events
// @Accept json
// @Produce json
// @Param limit query integer false "Number of events to retrieve" default(100)
// @Success 200 {object} APIResponse "List of unprocessed events"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /events/redemption-requested/unprocessed [get]
func GetUnprocessedRedemptionRequests(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 100
	}

	db := database.GetDB()
	events, err := models.GetUnprocessedRedemptionRequests(db, limit)
	if err != nil {
		InternalServerErrorWithDetails(c, "Failed to fetch unprocessed events", err.Error())
		return
	}

	SendSuccess(c, http.StatusOK, events, "")
}

// processSukukPurchase creates or updates an investment record based on the purchase event
func processSukukPurchase(tx *gorm.DB, event *models.SukukPurchased) error {
	// Find the sukuk by address
	var sukuk models.Sukuk
	err := tx.Where("token_address = ?", event.SukukAddress).First(&sukuk).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("sukuk not found for address: %s", event.SukukAddress)
		}
		return err
	}

	// Create investment record
	investment := models.Investment{
		SukukID:          sukuk.ID,
		InvestorAddress:  event.Buyer,
		InvestmentAmount: event.Amount,
		TokenAmount:      event.Amount,          // Assuming 1:1 for now, adjust based on your token economics
		TokenPrice:       "1000000000000000000", // 1e18 for 1:1 ratio, adjust as needed
		TxHash:           event.TxHash,
		LogIndex:         int(event.LogIndex),
		InvestmentDate:   event.Timestamp,
		Status:           models.InvestmentStatusActive,
	}

	return tx.Create(&investment).Error
}

// processRedemptionRequest creates a redemption record based on the redemption request event
func processRedemptionRequest(tx *gorm.DB, event *models.RedemptionRequested) error {
	// Find the sukuk by address
	var sukuk models.Sukuk
	err := tx.Where("token_address = ?", event.SukukAddress).First(&sukuk).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("sukuk not found for address: %s", event.SukukAddress)
		}
		return err
	}

	// Create redemption record
	redemption := models.Redemption{
		SukukID:          sukuk.ID,
		InvestorAddress:  event.User,
		TokenAmount:      event.Amount,
		RedemptionAmount: event.Amount, // Assuming 1:1 for now, adjust based on your token economics
		RequestTxHash:    event.TxHash,
		RequestLogIndex:  int(event.LogIndex),
		RequestDate:      event.Timestamp,
		Status:           models.RedemptionStatusRequested,
	}

	return tx.Create(&redemption).Error
}
