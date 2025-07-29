package services

import (
	"fmt"
	"regexp"
	"strings"

	"sukuk-be/internal/database"

	"gorm.io/gorm"
)

// IndexerTableService handles discovery of hash-prefixed indexer tables
type IndexerTableService struct {
	indexerDB *gorm.DB
}

// NewIndexerTableService creates a new table discovery service
func NewIndexerTableService() *IndexerTableService {
	return &IndexerTableService{}
}

// ConnectToIndexer connects to the Ponder indexer database
func (s *IndexerTableService) ConnectToIndexer() error {
	s.indexerDB = database.GetDB()
	return nil
}

// EventTableMapping defines the expected event table suffixes
var EventTableMapping = map[string]string{
	"sukuk_creation":         "sukuk_creation",
	"sukuk_purchase":         "sukuk_purchase", 
	"redemption_request":     "redemption_request",
	"redemption_approval":    "redemption_approval",
	"yield_distributed":      "yield_distribution",
	"yield_claimed":          "yield_claim",
	"snapshot_taken":         "snapshot_taken",
	"snapshot_criteria_update": "snapshot_criteria_update",
	"holder_addition":        "holder_addition",
	"holder_update":          "holder_update",
	"manager_update":         "manager_update",
	"vault_update":           "vault_update",
	"sale_status_change":     "sale_status_change",
	"sukuk_status_update":    "sukuk_status_update",
	"yield_deposit":          "yield_deposit",
	"yield_vault_manager_addition": "yield_vault_manager_addition",
	"yield_vault_manager_removal":  "yield_vault_manager_removal",
	"minter_addition":        "minter_addition",
	"minter_removal":         "minter_removal",
	"status_change":          "status_change",
}

// TableInfo represents discovered table information
type TableInfo struct {
	FullName    string // e.g., "f243__sukuk_creation"
	HashPrefix  string // e.g., "f243"
	EventType   string // e.g., "sukuk_creation"
	SchemaName  string // e.g., "public"
}

// DiscoverAllTables finds all hash-prefixed indexer tables
func (s *IndexerTableService) DiscoverAllTables() ([]TableInfo, error) {
	if s.indexerDB == nil {
		if err := s.ConnectToIndexer(); err != nil {
			return nil, err
		}
	}

	var tables []TableInfo
	
	// Query to get all table names that match hash-prefix pattern
	// Exclude _reorg tables as they are for blockchain reorganization handling
	query := `
		SELECT table_name, table_schema
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name ~ '^[a-f0-9]+__[a-z_]+$'
		AND table_name NOT LIKE '%_reorg__%'
		ORDER BY table_name DESC
	`
	
	rows, err := s.indexerDB.Raw(query).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	// Regex to parse hash-prefixed table names (handles both normal and _reorg variants)
	tableRegex := regexp.MustCompile(`^([a-f0-9]+)(?:_reorg)?__(.+)$`)
	
	for rows.Next() {
		var tableName, schemaName string
		if err := rows.Scan(&tableName, &schemaName); err != nil {
			continue
		}

		matches := tableRegex.FindStringSubmatch(tableName)
		if len(matches) == 3 {
			tables = append(tables, TableInfo{
				FullName:   tableName,
				HashPrefix: matches[1],
				EventType:  matches[2],
				SchemaName: schemaName,
			})
		}
	}

	return tables, nil
}

// GetLatestTableForEvent finds the latest table for a specific event type
// Uses max block number and row count to determine the most relevant table
func (s *IndexerTableService) GetLatestTableForEvent(eventType string) (string, error) {
	tables, err := s.DiscoverAllTables()
	if err != nil {
		return "", err
	}

	// Find all tables for this event type
	var eventTables []TableInfo
	for _, table := range tables {
		if table.EventType == eventType {
			eventTables = append(eventTables, table)
		}
	}

	if len(eventTables) == 0 {
		return "", fmt.Errorf("no tables found for event type: %s", eventType)
	}

	// If only one table, return it
	if len(eventTables) == 1 {
		return eventTables[0].FullName, nil
	}

	// Find the best table based on max block number and row count
	bestTable := eventTables[0]
	bestMaxBlock := int64(-1)
	bestRowCount := int64(-1)

	for _, table := range eventTables {
		// Get max block number for this table
		var maxBlock int64
		err := s.indexerDB.Raw(fmt.Sprintf("SELECT COALESCE(MAX(block_number), -1) FROM %s", table.FullName)).Scan(&maxBlock).Error
		if err != nil {
			// If query fails, skip this table
			continue
		}

		// Get row count for this table
		var rowCount int64
		err = s.indexerDB.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", table.FullName)).Scan(&rowCount).Error
		if err != nil {
			// If query fails, skip this table
			continue
		}

		// Priority logic:
		// 1. Table with higher max block number wins
		// 2. If equal block numbers, table with more rows wins
		// 3. If both equal, prefer non-reorg tables over reorg tables
		// 4. If still equal, use alphabetical ordering as fallback
		shouldUpdate := false
		isReorg := strings.Contains(table.FullName, "_reorg__")
		bestIsReorg := strings.Contains(bestTable.FullName, "_reorg__")
		
		if maxBlock > bestMaxBlock {
			shouldUpdate = true
		} else if maxBlock == bestMaxBlock && rowCount > bestRowCount {
			shouldUpdate = true
		} else if maxBlock == bestMaxBlock && rowCount == bestRowCount && bestIsReorg && !isReorg {
			// Prefer non-reorg table over reorg table
			shouldUpdate = true
		} else if maxBlock == bestMaxBlock && rowCount == bestRowCount && isReorg == bestIsReorg && table.FullName > bestTable.FullName {
			shouldUpdate = true
		}

		if shouldUpdate {
			bestTable = table
			bestMaxBlock = maxBlock
			bestRowCount = rowCount
		}
	}

	return bestTable.FullName, nil
}

// GetAllLatestTables returns a map of event type to latest table name
// Uses the same improved logic as GetLatestTableForEvent
func (s *IndexerTableService) GetAllLatestTables() (map[string]string, error) {
	tables, err := s.DiscoverAllTables()
	if err != nil {
		return nil, err
	}

	latestTables := make(map[string]string)
	
	// Group tables by event type
	eventGroups := make(map[string][]TableInfo)
	for _, table := range tables {
		eventGroups[table.EventType] = append(eventGroups[table.EventType], table)
	}

	// For each event type, find the best table using the same logic as GetLatestTableForEvent
	for eventType, eventTables := range eventGroups {
		if len(eventTables) == 1 {
			latestTables[eventType] = eventTables[0].FullName
			continue
		}

		// Find the best table based on max block number and row count
		bestTable := eventTables[0]
		bestMaxBlock := int64(-1)
		bestRowCount := int64(-1)

		for _, table := range eventTables {
			// Get max block number for this table
			var maxBlock int64
			err := s.indexerDB.Raw(fmt.Sprintf("SELECT COALESCE(MAX(block_number), -1) FROM %s", table.FullName)).Scan(&maxBlock).Error
			if err != nil {
				// If query fails, skip this table
				continue
			}

			// Get row count for this table
			var rowCount int64
			err = s.indexerDB.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", table.FullName)).Scan(&rowCount).Error
			if err != nil {
				// If query fails, skip this table
				continue
			}

			// Priority logic:
			// 1. Table with higher max block number wins
			// 2. If equal block numbers, table with more rows wins
			// 3. If both equal, prefer non-reorg tables over reorg tables
			// 4. If still equal, use alphabetical ordering as fallback
			shouldUpdate := false
			isReorg := strings.Contains(table.FullName, "_reorg__")
			bestIsReorg := strings.Contains(bestTable.FullName, "_reorg__")
			
			if maxBlock > bestMaxBlock {
				shouldUpdate = true
			} else if maxBlock == bestMaxBlock && rowCount > bestRowCount {
				shouldUpdate = true
			} else if maxBlock == bestMaxBlock && rowCount == bestRowCount && bestIsReorg && !isReorg {
				// Prefer non-reorg table over reorg table
				shouldUpdate = true
			} else if maxBlock == bestMaxBlock && rowCount == bestRowCount && isReorg == bestIsReorg && table.FullName > bestTable.FullName {
				shouldUpdate = true
			}

			if shouldUpdate {
				bestTable = table
				bestMaxBlock = maxBlock
				bestRowCount = rowCount
			}
		}

		latestTables[eventType] = bestTable.FullName
	}

	return latestTables, nil
}

// CheckTableExists verifies if a specific table exists
func (s *IndexerTableService) CheckTableExists(tableName string) (bool, error) {
	if s.indexerDB == nil {
		if err := s.ConnectToIndexer(); err != nil {
			return false, err
		}
	}

	var count int64
	err := s.indexerDB.Raw(`
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name = ?
	`, tableName).Scan(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetTablesByHashPrefix returns all tables with a specific hash prefix
func (s *IndexerTableService) GetTablesByHashPrefix(hashPrefix string) ([]TableInfo, error) {
	tables, err := s.DiscoverAllTables()
	if err != nil {
		return nil, err
	}

	var filteredTables []TableInfo
	for _, table := range tables {
		if table.HashPrefix == hashPrefix {
			filteredTables = append(filteredTables, table)
		}
	}

	return filteredTables, nil
}

// GetTableRowCount returns the number of rows in a table
func (s *IndexerTableService) GetTableRowCount(tableName string) (int64, error) {
	if s.indexerDB == nil {
		if err := s.ConnectToIndexer(); err != nil {
			return 0, err
		}
	}

	var count int64
	err := s.indexerDB.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count rows in table %s: %w", tableName, err)
	}

	return count, nil
}

// GetAvailableEventTypes returns all event types that have tables
func (s *IndexerTableService) GetAvailableEventTypes() ([]string, error) {
	latestTables, err := s.GetAllLatestTables()
	if err != nil {
		return nil, err
	}

	eventTypes := make([]string, 0, len(latestTables))
	for eventType := range latestTables {
		eventTypes = append(eventTypes, eventType)
	}

	return eventTypes, nil
}

// ValidateTableStructure checks if a table has expected columns for an event type
func (s *IndexerTableService) ValidateTableStructure(tableName string, eventType string) error {
	if s.indexerDB == nil {
		if err := s.ConnectToIndexer(); err != nil {
			return err
		}
	}

	// Get column names for the table
	var columns []string
	err := s.indexerDB.Raw(`
		SELECT column_name 
		FROM information_schema.columns 
		WHERE table_schema = 'public' 
		AND table_name = ?
		ORDER BY ordinal_position
	`, tableName).Pluck("column_name", &columns).Error

	if err != nil {
		return fmt.Errorf("failed to get columns for table %s: %w", tableName, err)
	}

	// Basic validation - all event tables should have these common fields
	requiredColumns := []string{"id", "timestamp", "block_number", "tx_hash"}
	
	for _, required := range requiredColumns {
		found := false
		for _, column := range columns {
			if strings.EqualFold(column, required) {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("table %s missing required column: %s", tableName, required)
		}
	}

	return nil
}