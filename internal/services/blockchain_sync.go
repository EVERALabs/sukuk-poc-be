package services

import (
	"fmt"
	"strconv"
	"time"

	"sukuk-be/internal/database"
	"sukuk-be/internal/logger"
	"sukuk-be/internal/models"

	"gorm.io/gorm"
)

// BlockchainEvent represents an event from the indexer database
type BlockchainEvent struct {
	ID              uint64                 `json:"id"`
	EventName       string                 `json:"event_name"`
	TxHash          string                 `json:"tx_hash"`
	LogIndex        int                    `json:"log_index"`
	BlockNumber     uint64                 `json:"block_number"`
	BlockTimestamp  time.Time              `json:"block_timestamp"`
	ContractAddress string                 `json:"contract_address"`
	EventData       map[string]interface{} `json:"event_data"`
	ChainID         int64                  `json:"chain_id"`
}

// BlockchainSyncService handles syncing blockchain events from indexer
type BlockchainSyncService struct {
	db              *gorm.DB
	lastProcessedID uint64
	syncInterval    time.Duration
	stopChan        chan bool
}

// NewBlockchainSyncService creates a new blockchain sync service
func NewBlockchainSyncService(syncInterval time.Duration) *BlockchainSyncService {
	return &BlockchainSyncService{
		db:           database.GetDB(),
		syncInterval: syncInterval,
		stopChan:     make(chan bool),
	}
}

// Start begins the sync process
func (s *BlockchainSyncService) Start() {
	logger.Info("Starting blockchain sync service")
	
	// Load last processed ID from database
	s.loadLastProcessedID()
	
	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			if err := s.syncNewEvents(); err != nil {
				logger.WithError(err).Error("Failed to sync blockchain events")
			}
		case <-s.stopChan:
			logger.Info("Stopping blockchain sync service")
			return
		}
	}
}

// Stop stops the sync service
func (s *BlockchainSyncService) Stop() {
	close(s.stopChan)
}

// loadLastProcessedID loads the last processed event ID from database
func (s *BlockchainSyncService) loadLastProcessedID() {
	var systemState models.SystemState
	if err := s.db.Where("key = ?", "last_processed_event_id").First(&systemState).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Initialize with 0 if not found
			s.lastProcessedID = 0
			s.saveLastProcessedID()
		} else {
			logger.WithError(err).Error("Failed to load last processed event ID")
		}
		return
	}
	
	if id, err := strconv.ParseUint(systemState.Value, 10, 64); err == nil {
		s.lastProcessedID = id
	}
}

// saveLastProcessedID saves the last processed event ID to database
func (s *BlockchainSyncService) saveLastProcessedID() {
	systemState := models.SystemState{
		Key:   "last_processed_event_id",
		Value: fmt.Sprintf("%d", s.lastProcessedID),
	}
	
	s.db.Save(&systemState)
}

// syncNewEvents queries for new events and processes them
func (s *BlockchainSyncService) syncNewEvents() error {
	// Query new events from indexer database (blockchain schema)
	var events []BlockchainEvent
	
	// Note: This assumes the indexer database is accessible
	// You might need to configure a separate database connection for blockchain schema
	err := s.db.Raw(`
		SELECT 
			id,
			event_name,
			tx_hash,
			log_index,
			block_number,
			block_timestamp,
			contract_address,
			event_data,
			chain_id
		FROM blockchain.events 
		WHERE id > ? 
		ORDER BY id ASC 
		LIMIT 1000
	`, s.lastProcessedID).Scan(&events).Error
	
	if err != nil {
		return fmt.Errorf("failed to query blockchain events: %w", err)
	}
	
	if len(events) == 0 {
		return nil // No new events
	}
	
	logger.WithField("count", len(events)).Info("Processing new blockchain events")
	
	// Process events in transaction
	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, event := range events {
			if err := s.processEvent(tx, &event); err != nil {
				logger.WithError(err).
					WithField("event_id", event.ID).
					WithField("event_name", event.EventName).
					Error("Failed to process event")
				// Continue processing other events
				continue
			}
			
			s.lastProcessedID = event.ID
		}
		
		// Save last processed ID
		s.saveLastProcessedID()
		return nil
	})
}

// processEvent processes a single blockchain event
func (s *BlockchainSyncService) processEvent(tx *gorm.DB, event *BlockchainEvent) error {
	switch event.EventName {
	case "SukukDeployed":
		return s.processSukukDeployedEvent(tx, event)
	case "Investment":
		return s.processInvestmentEvent(tx, event)
	case "YieldDistributed":
		return s.processYieldDistributedEvent(tx, event)
	case "YieldClaimed":
		return s.processYieldClaimedEvent(tx, event)
	case "RedemptionRequested":
		return s.processRedemptionRequestedEvent(tx, event)
	case "RedemptionApproved":
		return s.processRedemptionApprovedEvent(tx, event)
	case "RedemptionCompleted":
		return s.processRedemptionCompletedEvent(tx, event)
	case "RedemptionRejected":
		return s.processRedemptionRejectedEvent(tx, event)
	case "EmergencySuspended":
		return s.processEmergencySuspendedEvent(tx, event)
	default:
		logger.WithField("event_name", event.EventName).Warn("Unknown event type")
		return nil
	}
}

// processSukukDeployedEvent handles sukuk deployment events
func (s *BlockchainSyncService) processSukukDeployedEvent(tx *gorm.DB, event *BlockchainEvent) error {
	seriesName, _ := event.EventData["seriesName"].(string)
	tokenAddress, _ := event.EventData["tokenAddress"].(string)
	
	if seriesName == "" || tokenAddress == "" {
		return fmt.Errorf("missing required fields in SukukDeployed event")
	}
	
	// Find sukuk series by name and update token address
	var sukukSeries models.Sukuk
	if err := tx.Where("name = ? AND token_address = ''", seriesName).First(&sukukSeries).Error; err != nil {
		return fmt.Errorf("sukuk series not found for deployment: %w", err)
	}
	
	sukukSeries.TokenAddress = tokenAddress
	sukukSeries.Status = models.SukukStatusActive
	
	return tx.Save(&sukukSeries).Error
}

// processInvestmentEvent handles investment events
func (s *BlockchainSyncService) processInvestmentEvent(tx *gorm.DB, event *BlockchainEvent) error {
	investorAddress, _ := event.EventData["investor"].(string)
	sukukToken, _ := event.EventData["sukukToken"].(string)
	idrxAmount, _ := event.EventData["idrxAmount"].(string)
	tokenAmount, _ := event.EventData["tokenAmount"].(string)
	tokenPrice, _ := event.EventData["tokenPrice"].(string)
	previousOutstandingSupply, _ := event.EventData["previousOutstandingSupply"].(string)
	newOutstandingSupply, _ := event.EventData["newOutstandingSupply"].(string)
	
	if investorAddress == "" || sukukToken == "" || idrxAmount == "" || tokenAmount == "" || tokenPrice == "" {
		return fmt.Errorf("missing required fields in Investment event")
	}
	
	// Find sukuk series by token address
	var sukukSeries models.Sukuk
	if err := tx.Where("token_address = ?", sukukToken).First(&sukukSeries).Error; err != nil {
		return fmt.Errorf("sukuk series not found for token address %s: %w", sukukToken, err)
	}
	
	// Check if investment already exists (prevent duplicates)
	var existingInvestment models.Investment
	if err := tx.Where("tx_hash = ? AND log_index = ?", event.TxHash, event.LogIndex).First(&existingInvestment).Error; err == nil {
		return nil // Already processed
	}
	
	// Create investment record
	investment := models.Investment{
		SukukID:         sukukSeries.ID,
		InvestorAddress: investorAddress,
		InvestmentAmount: idrxAmount,
		TokenAmount:     tokenAmount,
		TokenPrice:      tokenPrice,
		Status:          models.InvestmentStatusActive,
		TxHash:          event.TxHash,
		LogIndex:        event.LogIndex,
		InvestmentDate:  event.BlockTimestamp,
	}
	
	// Update sukuk series outstanding supply if audit trail data is available
	if previousOutstandingSupply != "" && newOutstandingSupply != "" {
		// Log the state change for audit purposes
		logger.WithFields(map[string]interface{}{
			"sukuk_series_id": sukukSeries.ID,
			"previous_supply": previousOutstandingSupply,
			"new_supply":      newOutstandingSupply,
			"tx_hash":         event.TxHash,
		}).Info("Investment state change recorded")
		
		// Update the outstanding supply
		sukukSeries.OutstandingSupply = newOutstandingSupply
		if err := tx.Save(&sukukSeries).Error; err != nil {
			return fmt.Errorf("failed to update sukuk series outstanding supply: %w", err)
		}
	}
	
	return tx.Create(&investment).Error
}

// processYieldDistributedEvent handles yield distribution events
func (s *BlockchainSyncService) processYieldDistributedEvent(tx *gorm.DB, event *BlockchainEvent) error {
	distributionID, _ := event.EventData["distributionId"].(float64)
	sukukToken, _ := event.EventData["sukukToken"].(string)
	totalYieldAmount, _ := event.EventData["totalYieldAmount"].(string)
	periodStart, _ := event.EventData["periodStart"].(float64)
	periodEnd, _ := event.EventData["periodEnd"].(float64)
	yieldPerToken, _ := event.EventData["yieldPerToken"].(string)
	
	if sukukToken == "" || distributionID == 0 || totalYieldAmount == "" {
		return fmt.Errorf("missing required fields in YieldDistributed event")
	}
	
	// Find sukuk series
	var sukukSeries models.Sukuk
	if err := tx.Where("token_address = ?", sukukToken).First(&sukukSeries).Error; err != nil {
		return fmt.Errorf("sukuk series not found: %w", err)
	}
	
	// Get all active investments for this sukuk
	var investments []models.Investment
	if err := tx.Where("sukuk_id = ? AND status = ?", sukukSeries.ID, models.InvestmentStatusActive).Find(&investments).Error; err != nil {
		return fmt.Errorf("failed to get investments: %w", err)
	}
	
	// Create yield claims for each investment
	for _, investment := range investments {
		// Check if yield claim already exists for this distribution
		var existingClaim models.Yield
		if err := tx.Where("investment_id = ? AND distribution_id = ?", investment.ID, uint64(distributionID)).First(&existingClaim).Error; err == nil {
			continue // Already exists
		}
		
		// Calculate yield amount for this investment based on yieldPerToken
		var calculatedYield string
		if yieldPerToken != "" {
			// Calculate yield = tokenAmount * yieldPerToken / 1e18
			// This is a simplified calculation - in production you might want to use big.Int for precision
			calculatedYield = yieldPerToken // Placeholder - implement proper calculation
		} else {
			calculatedYield = "0"
		}
		
		// Convert Unix timestamps to time.Time
		periodStartTime := time.Unix(int64(periodStart), 0)
		periodEndTime := time.Unix(int64(periodEnd), 0)
		
		yieldClaim := models.Yield{
			SukukID:         sukukSeries.ID,
			InvestorAddress: investment.InvestorAddress,
			YieldAmount:     calculatedYield,
			PeriodStart:     periodStartTime,
			PeriodEnd:       periodEndTime,
			Status:          models.YieldStatusPending,
			DistributionDate: periodEndTime, // Use period end as distribution date
		}
		
		if err := tx.Create(&yieldClaim).Error; err != nil {
			return fmt.Errorf("failed to create yield claim: %w", err)
		}
	}
	
	return nil
}

// processYieldClaimedEvent handles yield claim events
func (s *BlockchainSyncService) processYieldClaimedEvent(tx *gorm.DB, event *BlockchainEvent) error {
	investorAddress, _ := event.EventData["investor"].(string)
	yieldAmount, _ := event.EventData["yieldAmount"].(string)
	fromDistribution, _ := event.EventData["fromDistribution"].(float64)
	toDistribution, _ := event.EventData["toDistribution"].(float64)
	
	if investorAddress == "" || yieldAmount == "" {
		return fmt.Errorf("missing required fields in YieldClaimed event")
	}
	
	// Update yield claims as claimed for the distribution range
	return tx.Model(&models.Yield{}).
		Where("investor_address = ? AND distribution_id >= ? AND distribution_id <= ? AND status = ?",
			investorAddress, uint64(fromDistribution), uint64(toDistribution), models.YieldStatusPending).
		Updates(map[string]interface{}{
			"status":           models.YieldStatusClaimed,
			"claimed_at":       &event.BlockTimestamp,
			"transaction_hash": event.TxHash,
			"block_number":     event.BlockNumber,
		}).Error
}

// Placeholder implementations for other event handlers
func (s *BlockchainSyncService) processRedemptionRequestedEvent(tx *gorm.DB, event *BlockchainEvent) error {
	// TODO: Implement redemption request processing
	return nil
}

func (s *BlockchainSyncService) processRedemptionApprovedEvent(tx *gorm.DB, event *BlockchainEvent) error {
	// TODO: Implement redemption approval processing
	return nil
}

func (s *BlockchainSyncService) processRedemptionCompletedEvent(tx *gorm.DB, event *BlockchainEvent) error {
	redemptionID, _ := event.EventData["redemptionId"].(string)
	investor, _ := event.EventData["investor"].(string)
	tokensBurned, _ := event.EventData["tokensBurned"].(string)
	idrxPaid, _ := event.EventData["idrxPaid"].(string)
	previousTokenBalance, _ := event.EventData["previousTokenBalance"].(string)
	newTokenBalance, _ := event.EventData["newTokenBalance"].(string)
	previousOutstandingSupply, _ := event.EventData["previousOutstandingSupply"].(string)
	newOutstandingSupply, _ := event.EventData["newOutstandingSupply"].(string)
	
	if redemptionID == "" || investor == "" || tokensBurned == "" || idrxPaid == "" {
		return fmt.Errorf("missing required fields in RedemptionCompleted event")
	}
	
	// Find redemption by ID (assuming redemptionId is stored in external_id field)
	var redemption models.Redemption
	if err := tx.Where("external_id = ?", redemptionID).First(&redemption).Error; err != nil {
		return fmt.Errorf("redemption not found for ID %s: %w", redemptionID, err)
	}
	
	// Update redemption status
	now := event.BlockTimestamp
	redemption.Status = models.RedemptionStatusCompleted
	redemption.CompletedAt = &now
	redemption.CompleteTxHash = event.TxHash
	redemption.CompleteLogIndex = &event.LogIndex
	
	if err := tx.Save(&redemption).Error; err != nil {
		return fmt.Errorf("failed to update redemption: %w", err)
	}
	
	// Note: Investment status update removed as Redemption model doesn't have InvestmentID field
	// The relationship between redemption and investment should be handled differently
	// if needed based on the actual model relationships
	
	// Update outstanding supply if audit data is available
	if previousOutstandingSupply != "" && newOutstandingSupply != "" {
		var sukuk models.Sukuk
		if err := tx.First(&sukuk, redemption.SukukID).Error; err == nil {
			sukuk.OutstandingSupply = newOutstandingSupply
			tx.Save(&sukuk)
			
			// Log audit trail
			logger.WithFields(map[string]interface{}{
				"redemption_id":    redemptionID,
				"previous_supply":  previousOutstandingSupply,
				"new_supply":       newOutstandingSupply,
				"previous_balance": previousTokenBalance,
				"new_balance":      newTokenBalance,
				"tx_hash":          event.TxHash,
			}).Info("Redemption state change recorded")
		}
	}
	
	return nil
}

func (s *BlockchainSyncService) processRedemptionRejectedEvent(tx *gorm.DB, event *BlockchainEvent) error {
	// TODO: Implement redemption rejection processing
	return nil
}

func (s *BlockchainSyncService) processEmergencySuspendedEvent(tx *gorm.DB, event *BlockchainEvent) error {
	// TODO: Implement emergency suspension processing
	return nil
}