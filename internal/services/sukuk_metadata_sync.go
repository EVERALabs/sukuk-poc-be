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

// SukukMetadataSyncService handles syncing sukuk data from indexer to metadata table
type SukukMetadataSyncService struct {
	db              *gorm.DB
	syncInterval    time.Duration
	stopChan        chan bool
	lastProcessedID uint64
}

// SukukCreationEvent represents a sukuk creation event from the indexer
type SukukCreationEvent struct {
	ID                string `gorm:"column:id;primaryKey"`
	TokenAddress      string `gorm:"column:token_address"`
	Name              string `gorm:"column:name"`
	Symbol            string `gorm:"column:symbol"`
	Issuer            string `gorm:"column:issuer"`
	Manager           string `gorm:"column:manager"`
	MaxSupply         string `gorm:"column:max_supply"`
	MaturityTimestamp int64  `gorm:"column:maturity_timestamp"`
	BlockNumber       int64  `gorm:"column:block_number"`
	TxHash            string `gorm:"column:tx_hash"`
	Timestamp         int64  `gorm:"column:timestamp"`
}


// NewSukukMetadataSyncService creates a new metadata sync service
func NewSukukMetadataSyncService(syncInterval time.Duration) *SukukMetadataSyncService {
	return &SukukMetadataSyncService{
		db:           database.GetDB(),
		syncInterval: syncInterval,
		stopChan:     make(chan bool),
	}
}

// Start begins the sync process
func (s *SukukMetadataSyncService) Start() {
	logger.Info("Starting sukuk metadata sync service")
	
	// Load last processed ID
	s.loadLastProcessedID()
	
	// Start sync loop
	go s.syncLoop()
}

// Stop stops the sync service
func (s *SukukMetadataSyncService) Stop() {
	logger.Info("Stopping sukuk metadata sync service")
	close(s.stopChan)
}

// syncLoop runs the main sync loop
func (s *SukukMetadataSyncService) syncLoop() {
	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()

	// Run immediately on start
	s.syncEvents()

	for {
		select {
		case <-ticker.C:
			s.syncEvents()
		case <-s.stopChan:
			return
		}
	}
}

// loadLastProcessedID loads the last processed event ID from system state
func (s *SukukMetadataSyncService) loadLastProcessedID() {
	var state models.SystemState
	result := s.db.Where("key = ?", "sukuk_metadata_last_event_id").First(&state)
	
	if result.Error == nil {
		// Parse the string value to uint64
		lastID, err := strconv.ParseUint(state.Value, 10, 64)
		if err == nil {
			s.lastProcessedID = lastID
			logger.WithField("last_processed_id", s.lastProcessedID).Info("Loaded last processed event ID")
		} else {
			logger.WithError(err).Error("Failed to parse last processed ID")
			s.lastProcessedID = 0
		}
	} else {
		logger.Info("No previous sync state found, starting from beginning")
		s.lastProcessedID = 0
	}
}

// saveLastProcessedID saves the last processed event ID
func (s *SukukMetadataSyncService) saveLastProcessedID(eventID uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var state models.SystemState
		result := tx.Where("key = ?", "sukuk_metadata_last_event_id").First(&state)
		
		if result.Error == gorm.ErrRecordNotFound {
			// Create new state
			state = models.SystemState{
				Key:   "sukuk_metadata_last_event_id",
				Value: strconv.FormatUint(eventID, 10),
			}
			return tx.Create(&state).Error
		}
		
		// Update existing state
		state.Value = strconv.FormatUint(eventID, 10)
		return tx.Save(&state).Error
	})
}

// syncEvents fetches and processes new events from the indexer
func (s *SukukMetadataSyncService) syncEvents() {
	logger.Debug("Starting metadata sync cycle")
	
	// First, find the most recent sukuk creation table
	tableName, err := s.FindLatestSukukCreationTable()
	if err != nil {
		logger.WithError(err).Error("Failed to find sukuk creation table")
		return
	}
	
	if tableName == "" {
		logger.Debug("No sukuk creation tables found")
		return
	}
	
	logger.WithField("table_name", tableName).Debug("Using sukuk creation table")
	
	// Query new events from the latest table
	var events []SukukCreationEvent
	result := s.db.Table(tableName).
		Order("timestamp DESC").  // Use timestamp for ordering instead of hex ID
		Limit(100).
		Find(&events)
	
	if result.Error != nil {
		logger.WithError(result.Error).Error("Failed to fetch events from indexer")
		return
	}
	
	if len(events) == 0 {
		logger.Debug("No sukuk events to process")
		return
	}
	
	logger.WithField("count", len(events)).Info("Processing sukuk metadata events")
	
	// Process each event
	for _, event := range events {
		// Check if we already have this sukuk to avoid duplicates
		var existing models.SukukMetadata
		existsResult := s.db.Where("contract_address = ?", event.TokenAddress).First(&existing)
		
		if existsResult.Error == nil {
			logger.WithField("contract_address", event.TokenAddress).Debug("Sukuk already exists, skipping")
			continue
		}
		
		if err := s.processEvent(&event); err != nil {
			logger.WithError(err).WithField("event_id", event.ID).Error("Failed to process event")
			continue
		}
	}
}

// processEvent processes a single sukuk creation event
func (s *SukukMetadataSyncService) processEvent(event *SukukCreationEvent) error {
	logger.WithFields(map[string]interface{}{
		"symbol":   event.Symbol,
		"name":     event.Name,
		"tx_hash":  event.TxHash,
		"event_id": event.ID,
	}).Info("Processing sukuk creation event")
	
	// Check if metadata already exists (using token_address as unique identifier)
	var existing models.SukukMetadata
	result := s.db.Where("contract_address = ?", event.TokenAddress).First(&existing)
	
	if result.Error == nil {
		// Update existing metadata
		return s.updateSukukMetadata(&existing, event)
	}
	
	// Create new metadata
	return s.createSukukMetadata(event)
}

// createSukukMetadata creates new sukuk metadata from blockchain event
func (s *SukukMetadataSyncService) createSukukMetadata(event *SukukCreationEvent) error {
	// Map blockchain data to metadata fields
	metadata := models.SukukMetadata{
		// Onchain data
		ContractAddress: event.TokenAddress,
		TokenID:         0, // Will be set when we have the actual token ID
		TransactionHash: event.TxHash,
		BlockNumber:     event.BlockNumber,
		
		// Basic info from event
		SukukCode:  event.Symbol,
		SukukTitle: event.Name,
		Status:     "berlangsung", // Default status
		
		// Financial info (will need offchain data for complete info)
		KuotaNasional: s.parseAmount(event.MaxSupply),
		
		// Dates
		JatuhTempo: time.Unix(event.MaturityTimestamp, 0),
		
		// Default values - these need to be updated with offchain data
		PenerimaanKupon: "Bulanan", // Default, update from offchain
		TipeKupon:       "Fixed Rate", // Default, update from offchain
		
		// Additional fields from indexer
		SukukDeskripsi: fmt.Sprintf("Sukuk issued by %s, managed by %s", event.Issuer, event.Manager),
		
		// Not ready until offchain data is added
		MetadataReady: false,
	}
	
	// Save to database
	if err := s.db.Create(&metadata).Error; err != nil {
		return fmt.Errorf("failed to create sukuk metadata: %w", err)
	}
	
	logger.WithField("sukuk_code", metadata.SukukCode).Info("Created new sukuk metadata from blockchain event")
	return nil
}

// updateSukukMetadata updates existing metadata with new blockchain data
func (s *SukukMetadataSyncService) updateSukukMetadata(metadata *models.SukukMetadata, event *SukukCreationEvent) error {
	// Update onchain data if changed
	metadata.TransactionHash = event.TxHash
	metadata.BlockNumber = event.BlockNumber
	
	// Update basic info
	metadata.SukukTitle = event.Name
	metadata.SukukCode = event.Symbol
	metadata.JatuhTempo = time.Unix(event.MaturityTimestamp, 0)
	metadata.KuotaNasional = s.parseAmount(event.MaxSupply)
	
	// Save updates
	if err := s.db.Save(metadata).Error; err != nil {
		return fmt.Errorf("failed to update sukuk metadata: %w", err)
	}
	
	logger.WithField("sukuk_code", metadata.SukukCode).Info("Updated sukuk metadata from blockchain event")
	return nil
}


// parseAmount parses amount string to float64, handling large numbers
func (s *SukukMetadataSyncService) parseAmount(amount string) float64 {
	if amount == "" {
		return 0
	}
	
	// Try to parse as regular number first
	if val, err := strconv.ParseFloat(amount, 64); err == nil {
		// Convert from wei to normal units if the number is very large (likely wei)
		if val > 1e18 {
			return val / 1e18 // Convert from wei to ether/token units
		}
		return val
	}
	
	// If parsing fails, try to extract numeric part
	var result float64
	fmt.Sscanf(amount, "%f", &result)
	
	// Convert from wei if very large
	if result > 1e18 {
		return result / 1e18
	}
	
	return result
}

// SyncSpecificSukuk manually syncs a specific sukuk by contract address
func (s *SukukMetadataSyncService) SyncSpecificSukuk(tokenID int64, contractAddress string) error {
	// Find all sukuk creation tables and search through them
	tables, err := s.FindAllSukukCreationTables()
	if err != nil {
		return fmt.Errorf("failed to find sukuk creation tables: %w", err)
	}
	
	// Search through all tables for the specific contract address
	for _, tableName := range tables {
		var event SukukCreationEvent
		result := s.db.Table(tableName).
			Where("token_address = ?", contractAddress).
			First(&event)
		
		if result.Error == nil {
			// Found the event, process it
			logger.WithFields(map[string]interface{}{
				"table_name": tableName,
				"contract_address": contractAddress,
			}).Info("Found sukuk in table")
			return s.processEvent(&event)
		}
	}
	
	return fmt.Errorf("sukuk with contract address %s not found in any indexer table", contractAddress)
}

// FindLatestSukukCreationTable finds the most recent sukuk creation table
func (s *SukukMetadataSyncService) FindLatestSukukCreationTable() (string, error) {
	// Query to find all tables that match the sukuk creation pattern
	var tables []struct {
		TableName string `gorm:"column:table_name"`
	}
	
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name LIKE '%sukuk_creation' 
		ORDER BY table_name DESC 
		LIMIT 1`
	
	result := s.db.Raw(query).Scan(&tables)
	if result.Error != nil {
		return "", fmt.Errorf("failed to query sukuk creation tables: %w", result.Error)
	}
	
	if len(tables) == 0 {
		return "", nil
	}
	
	return tables[0].TableName, nil
}

// FindAllSukukCreationTables finds all sukuk creation tables and returns them with their creation info
func (s *SukukMetadataSyncService) FindAllSukukCreationTables() ([]string, error) {
	var tables []struct {
		TableName string `gorm:"column:table_name"`
	}
	
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name LIKE '%sukuk_creation' 
		ORDER BY table_name DESC`
	
	result := s.db.Raw(query).Scan(&tables)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to query sukuk creation tables: %w", result.Error)
	}
	
	var tableNames []string
	for _, table := range tables {
		tableNames = append(tableNames, table.TableName)
	}
	
	return tableNames, nil
}

// parseEventID converts hex event ID to uint64 for tracking
func (s *SukukMetadataSyncService) parseEventID(hexID string) (uint64, error) {
	// Remove 0x prefix if present
	if len(hexID) > 2 && hexID[:2] == "0x" {
		hexID = hexID[2:]
	}
	
	return strconv.ParseUint(hexID, 16, 64)
}