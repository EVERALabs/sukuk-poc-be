package services

import (
	"fmt"
	"time"

	"sukuk-be/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type IndexerQueryService struct {
	indexerDB *gorm.DB
}

// NewIndexerQueryService creates a new service to query indexer database
func NewIndexerQueryService() *IndexerQueryService {
	return &IndexerQueryService{}
}

// ConnectToIndexer connects to the Ponder indexer database
func (s *IndexerQueryService) ConnectToIndexer() error {
	// Get indexer database URL from environment or config
	indexerDBURL := "postgresql://postgres:postgres@localhost:5432/sukuk_poc_new" // Default from indexer .env
	
	db, err := gorm.Open(postgres.Open(indexerDBURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to indexer database: %w", err)
	}
	
	s.indexerDB = db
	return nil
}

// GetLatestActivities queries the indexer database directly for latest activities
func (s *IndexerQueryService) GetLatestActivities(sukukAddress string, limit int) ([]models.ActivityEvent, error) {
	if s.indexerDB == nil {
		if err := s.ConnectToIndexer(); err != nil {
			return nil, err
		}
	}

	if limit == 0 {
		limit = 10
	}

	var activities []models.ActivityEvent

	// Query sukuk_purchase table directly from indexer
	var purchases []IndexerSukukPurchase
	err := s.indexerDB.Table("sukuk_purchase").
		Where("sukuk_address = ?", sukukAddress).
		Order("timestamp DESC").
		Limit(limit).
		Find(&purchases).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query sukuk purchases: %w", err)
	}

	// Query redemption_request table directly from indexer
	var redemptions []IndexerRedemptionRequest
	err = s.indexerDB.Table("redemption_request").
		Where("sukuk_address = ?", sukukAddress).
		Order("timestamp DESC").
		Limit(limit).
		Find(&redemptions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query redemption requests: %w", err)
	}

	// Convert to ActivityEvent and merge
	for _, p := range purchases {
		activities = append(activities, models.ActivityEvent{
			Type:         "purchase",
			Address:      p.Buyer,
			Amount:       p.Amount,
			TxHash:       p.TxHash,
			Timestamp:    time.Unix(p.Timestamp, 0),
			SukukAddress: p.SukukAddress,
		})
	}

	for _, r := range redemptions {
		activities = append(activities, models.ActivityEvent{
			Type:         "redemption_request",
			Address:      r.User,
			Amount:       r.Amount,
			TxHash:       r.TxHash,
			Timestamp:    time.Unix(r.Timestamp, 0),
			SukukAddress: r.SukukAddress,
		})
	}

	// Sort by timestamp descending and limit
	for i := 0; i < len(activities)-1; i++ {
		for j := i + 1; j < len(activities); j++ {
			if activities[i].Timestamp.Before(activities[j].Timestamp) {
				activities[i], activities[j] = activities[j], activities[i]
			}
		}
	}

	if len(activities) > limit {
		activities = activities[:limit]
	}

	return activities, nil
}

// GetSukukPurchases gets purchase events for a specific sukuk
func (s *IndexerQueryService) GetSukukPurchases(sukukAddress string, limit int) ([]IndexerSukukPurchase, error) {
	if s.indexerDB == nil {
		if err := s.ConnectToIndexer(); err != nil {
			return nil, err
		}
	}

	var purchases []IndexerSukukPurchase
	query := s.indexerDB.Table("sukuk_purchase").
		Order("timestamp DESC")

	if sukukAddress != "" {
		query = query.Where("sukuk_address = ?", sukukAddress)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&purchases).Error
	return purchases, err
}

// GetRedemptionRequests gets redemption request events for a specific sukuk
func (s *IndexerQueryService) GetRedemptionRequests(sukukAddress string, limit int) ([]IndexerRedemptionRequest, error) {
	if s.indexerDB == nil {
		if err := s.ConnectToIndexer(); err != nil {
			return nil, err
		}
	}

	var redemptions []IndexerRedemptionRequest
	query := s.indexerDB.Table("redemption_request").
		Order("timestamp DESC")

	if sukukAddress != "" {
		query = query.Where("sukuk_address = ?", sukukAddress)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&redemptions).Error
	return redemptions, err
}

// GetActivitiesByAddress gets all activities (purchases + redemptions) for a specific address
func (s *IndexerQueryService) GetActivitiesByAddress(userAddress string, limit int) ([]models.ActivityEvent, error) {
	if s.indexerDB == nil {
		if err := s.ConnectToIndexer(); err != nil {
			return nil, err
		}
	}

	if limit == 0 {
		limit = 50 // Default higher limit for user history
	}

	var activities []models.ActivityEvent

	// Query sukuk_purchase table for user's purchases
	var purchases []IndexerSukukPurchase
	err := s.indexerDB.Table("sukuk_purchase").
		Where("buyer = ?", userAddress).
		Order("timestamp DESC").
		Limit(limit).
		Find(&purchases).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query user purchases: %w", err)
	}

	// Query redemption_request table for user's redemptions
	var redemptions []IndexerRedemptionRequest
	err = s.indexerDB.Table("redemption_request").
		Where("user = ?", userAddress).
		Order("timestamp DESC").
		Limit(limit).
		Find(&redemptions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query user redemptions: %w", err)
	}

	// Convert purchases to ActivityEvent
	for _, p := range purchases {
		activities = append(activities, models.ActivityEvent{
			Type:         "purchase",
			Address:      p.Buyer,
			Amount:       p.Amount,
			TxHash:       p.TxHash,
			Timestamp:    time.Unix(p.Timestamp, 0),
			SukukAddress: p.SukukAddress,
		})
	}

	// Convert redemptions to ActivityEvent
	for _, r := range redemptions {
		activities = append(activities, models.ActivityEvent{
			Type:         "redemption_request",
			Address:      r.User,
			Amount:       r.Amount,
			TxHash:       r.TxHash,
			Timestamp:    time.Unix(r.Timestamp, 0),
			SukukAddress: r.SukukAddress,
		})
	}

	// Sort by timestamp descending and limit
	for i := 0; i < len(activities)-1; i++ {
		for j := i + 1; j < len(activities); j++ {
			if activities[i].Timestamp.Before(activities[j].Timestamp) {
				activities[i], activities[j] = activities[j], activities[i]
			}
		}
	}

	if len(activities) > limit {
		activities = activities[:limit]
	}

	return activities, nil
}

// GetSukukOwnedByAddress gets unique sukuk addresses that a user has purchased
func (s *IndexerQueryService) GetSukukOwnedByAddress(userAddress string) ([]string, error) {
	if s.indexerDB == nil {
		if err := s.ConnectToIndexer(); err != nil {
			return nil, err
		}
	}

	var sukukAddresses []string
	
	// Query for distinct sukuk addresses from purchases
	err := s.indexerDB.Table("sukuk_purchase").
		Select("DISTINCT sukuk_address").
		Where("buyer = ?", userAddress).
		Pluck("sukuk_address", &sukukAddresses).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to query owned sukuk addresses: %w", err)
	}

	return sukukAddresses, nil
}

// Data structures for indexer events (matching Ponder schema)
type IndexerSukukPurchase struct {
	ID           string `gorm:"column:id"`
	Buyer        string `gorm:"column:buyer"`
	SukukAddress string `gorm:"column:sukuk_address"`
	PaymentToken string `gorm:"column:payment_token"`
	Amount       string `gorm:"column:amount"`
	BlockNumber  int64  `gorm:"column:block_number"`
	TxHash       string `gorm:"column:tx_hash"`
	Timestamp    int64  `gorm:"column:timestamp"`
}

type IndexerRedemptionRequest struct {
	ID           string `gorm:"column:id"`
	User         string `gorm:"column:user"`
	SukukAddress string `gorm:"column:sukuk_address"`
	Amount       string `gorm:"column:amount"`
	PaymentToken string `gorm:"column:payment_token"`
	BlockNumber  int64  `gorm:"column:block_number"`
	TxHash       string `gorm:"column:tx_hash"`
	Timestamp    int64  `gorm:"column:timestamp"`
}