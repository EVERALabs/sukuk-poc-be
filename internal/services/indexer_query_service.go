package services

import (
	"fmt"
	"sort"
	"time"

	"sukuk-be/internal/database"
	"sukuk-be/internal/models"
	"sukuk-be/internal/utils"

	"gorm.io/gorm"
)

type IndexerQueryService struct {
	indexerDB    *gorm.DB
	tableService *IndexerTableService
}

// NewIndexerQueryService creates a new service to query indexer database
func NewIndexerQueryService() *IndexerQueryService {
	return &IndexerQueryService{
		tableService: NewIndexerTableService(),
	}
}

// ConnectToIndexer connects to the Ponder indexer database (same as main database)
func (s *IndexerQueryService) ConnectToIndexer() error {
	// Use the same database connection as the main application
	s.indexerDB = database.GetDB()
	// Also initialize the table service
	if s.tableService == nil {
		s.tableService = NewIndexerTableService()
	}
	return s.tableService.ConnectToIndexer()
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

	// Get latest table names using dynamic discovery
	purchaseTable, err := s.tableService.GetLatestTableForEvent("sukuk_purchase")
	if err != nil {
		return nil, fmt.Errorf("failed to find sukuk_purchase table: %w", err)
	}

	redemptionTable, err := s.tableService.GetLatestTableForEvent("redemption_request")
	if err != nil {
		return nil, fmt.Errorf("failed to find redemption_request table: %w", err)
	}

	// Query sukuk_purchase table directly from indexer
	var purchases []IndexerSukukPurchase
	err = s.indexerDB.Table(purchaseTable).
		Where("sukuk_address = ?", sukukAddress).
		Order("timestamp DESC").
		Limit(limit).
		Find(&purchases).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query sukuk purchases from %s: %w", purchaseTable, err)
	}

	// Query redemption_request table directly from indexer
	var redemptions []IndexerRedemptionRequest
	err = s.indexerDB.Table(redemptionTable).
		Where("sukuk_address = ?", sukukAddress).
		Order("timestamp DESC").
		Limit(limit).
		Find(&redemptions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query redemption requests from %s: %w", redemptionTable, err)
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

	// Enrich activities with sukuk metadata
	enrichedActivities, err := s.enrichActivitiesWithSukukMetadata(activities)
	if err != nil {
		return nil, fmt.Errorf("failed to enrich activities with sukuk metadata: %w", err)
	}

	return enrichedActivities, nil
}

// GetSukukPurchases gets purchase events for a specific sukuk
func (s *IndexerQueryService) GetSukukPurchases(sukukAddress string, limit int) ([]IndexerSukukPurchase, error) {
	if s.indexerDB == nil {
		if err := s.ConnectToIndexer(); err != nil {
			return nil, err
		}
	}

	// Get latest table name using dynamic discovery
	purchaseTable, err := s.tableService.GetLatestTableForEvent("sukuk_purchase")
	if err != nil {
		return nil, fmt.Errorf("failed to find sukuk_purchase table: %w", err)
	}

	var purchases []IndexerSukukPurchase
	query := s.indexerDB.Table(purchaseTable).
		Order("timestamp DESC")

	if sukukAddress != "" {
		query = query.Where("sukuk_address = ?", sukukAddress)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	err = query.Find(&purchases).Error
	return purchases, err
}

// GetRedemptionRequests gets redemption request events for a specific sukuk
func (s *IndexerQueryService) GetRedemptionRequests(sukukAddress string, limit int) ([]IndexerRedemptionRequest, error) {
	if s.indexerDB == nil {
		if err := s.ConnectToIndexer(); err != nil {
			return nil, err
		}
	}

	// Get latest table name using dynamic discovery
	redemptionTable, err := s.tableService.GetLatestTableForEvent("redemption_request")
	if err != nil {
		return nil, fmt.Errorf("failed to find redemption_request table: %w", err)
	}

	var redemptions []IndexerRedemptionRequest
	query := s.indexerDB.Table(redemptionTable).
		Order("timestamp DESC")

	if sukukAddress != "" {
		query = query.Where("sukuk_address = ?", sukukAddress)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	err = query.Find(&redemptions).Error
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

	// Get latest table names using dynamic discovery
	purchaseTable, err := s.tableService.GetLatestTableForEvent("sukuk_purchase")
	if err != nil {
		return nil, fmt.Errorf("failed to find sukuk_purchase table: %w", err)
	}

	redemptionTable, err := s.tableService.GetLatestTableForEvent("redemption_request")
	if err != nil {
		return nil, fmt.Errorf("failed to find redemption_request table: %w", err)
	}

	// Query sukuk_purchase table for user's purchases
	var purchases []IndexerSukukPurchase
	err = s.indexerDB.Table(purchaseTable).
		Where("buyer = ?", userAddress).
		Order("timestamp DESC").
		Limit(limit).
		Find(&purchases).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query user purchases from %s: %w", purchaseTable, err)
	}

	// Query redemption_request table for user's redemptions
	var redemptions []IndexerRedemptionRequest
	err = s.indexerDB.Table(redemptionTable).
		Where("user = ?", userAddress).
		Order("timestamp DESC").
		Limit(limit).
		Find(&redemptions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query user redemptions from %s: %w", redemptionTable, err)
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

	// Enrich activities with sukuk metadata
	enrichedActivities, err := s.enrichActivitiesWithSukukMetadata(activities)
	if err != nil {
		return nil, fmt.Errorf("failed to enrich activities with sukuk metadata: %w", err)
	}

	return enrichedActivities, nil
}

// GetSukukOwnedByAddress gets unique sukuk addresses that a user has purchased
func (s *IndexerQueryService) GetSukukOwnedByAddress(userAddress string) ([]string, error) {
	if s.indexerDB == nil {
		if err := s.ConnectToIndexer(); err != nil {
			return nil, err
		}
	}

	// Get latest table name using dynamic discovery
	purchaseTable, err := s.tableService.GetLatestTableForEvent("sukuk_purchase")
	if err != nil {
		return nil, fmt.Errorf("failed to find sukuk_purchase table: %w", err)
	}

	var sukukAddresses []string
	
	// Query for distinct sukuk addresses from purchases
	err = s.indexerDB.Table(purchaseTable).
		Select("DISTINCT sukuk_address").
		Where("buyer = ?", userAddress).
		Pluck("sukuk_address", &sukukAddresses).Error
	
	if err != nil {
		return nil, fmt.Errorf("failed to query owned sukuk addresses from %s: %w", purchaseTable, err)
	}

	return sukukAddresses, nil
}

// GetUnclaimedDistributionIds returns distribution IDs that a user can claim for a specific sukuk
func (s *IndexerQueryService) GetUnclaimedDistributionIds(userAddress string, sukukAddress string) ([]int64, error) {
	if s.indexerDB == nil {
		if err := s.ConnectToIndexer(); err != nil {
			return nil, err
		}
	}

	// Get latest table names using dynamic discovery
	distributedTable, err := s.tableService.GetLatestTableForEvent("yield_distributed")
	if err != nil {
		return nil, fmt.Errorf("failed to find yield_distributed table: %w", err)
	}

	claimedTable, err := s.tableService.GetLatestTableForEvent("yield_claimed")
	if err != nil {
		return nil, fmt.Errorf("failed to find yield_claimed table: %w", err)
	}

	// Get all yield distributions for this sukuk
	var distributions []IndexerYieldDistributed
	err = s.indexerDB.Table(distributedTable).
		Where("sukuk_address = ?", sukukAddress).
		Order("distribution_id ASC").
		Find(&distributions).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query yield distributions from %s: %w", distributedTable, err)
	}

	// Get all yield claims by this user for this sukuk
	var claims []IndexerYieldClaimed
	err = s.indexerDB.Table(claimedTable).
		Where("user = ? AND sukuk_address = ?", userAddress, sukukAddress).
		Find(&claims).Error
	if err != nil {
		return nil, fmt.Errorf("failed to query yield claims from %s: %w", claimedTable, err)
	}

	// Create a map of claimed distribution IDs
	claimedMap := make(map[int64]bool)
	for _, claim := range claims {
		claimedMap[claim.DistributionId] = true
	}

	// Filter out claimed distributions
	var unclaimedIds []int64
	for _, dist := range distributions {
		if !claimedMap[dist.DistributionId] {
			unclaimedIds = append(unclaimedIds, dist.DistributionId)
		}
	}

	return unclaimedIds, nil
}

// enrichActivitiesWithSukukMetadata enriches activities with sukuk metadata (code and title)
func (s *IndexerQueryService) enrichActivitiesWithSukukMetadata(activities []models.ActivityEvent) ([]models.ActivityEvent, error) {
	if len(activities) == 0 {
		return activities, nil
	}

	// Extract unique sukuk addresses
	sukukAddressMap := make(map[string]bool)
	for _, activity := range activities {
		sukukAddressMap[activity.SukukAddress] = true
	}

	// Convert to slice for batch query
	sukukAddresses := make([]string, 0, len(sukukAddressMap))
	for address := range sukukAddressMap {
		sukukAddresses = append(sukukAddresses, address)
	}

	// Batch fetch sukuk metadata
	var sukukMetadata []models.SukukMetadata
	err := s.indexerDB.Where("contract_address IN ?", sukukAddresses).Find(&sukukMetadata).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sukuk metadata: %w", err)
	}

	// Create lookup map for quick access
	metadataMap := make(map[string]models.SukukMetadata)
	for _, metadata := range sukukMetadata {
		metadataMap[metadata.ContractAddress] = metadata
	}

	// Enrich activities with metadata
	enrichedActivities := make([]models.ActivityEvent, len(activities))
	for i, activity := range activities {
		enrichedActivities[i] = activity
		if metadata, exists := metadataMap[activity.SukukAddress]; exists {
			enrichedActivities[i].SukukCode = metadata.SukukCode
			enrichedActivities[i].SukukTitle = metadata.SukukTitle
		}
	}

	return enrichedActivities, nil
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
	TotalSupply  string `gorm:"column:total_supply"`
	BlockNumber  int64  `gorm:"column:block_number"`
	TxHash       string `gorm:"column:tx_hash"`
	Timestamp    int64  `gorm:"column:timestamp"`
}

type IndexerRedemptionApproval struct {
	ID           string `gorm:"column:id"`
	User         string `gorm:"column:user"`
	SukukAddress string `gorm:"column:sukuk_address"`
	Amount       string `gorm:"column:amount"`
	TotalSupply  string `gorm:"column:total_supply"`
	BlockNumber  int64  `gorm:"column:block_number"`
	TxHash       string `gorm:"column:tx_hash"`
	Timestamp    int64  `gorm:"column:timestamp"`
}

// Additional indexer event structures for yield and portfolio management
type IndexerYieldDistributed struct {
	ID             string `gorm:"column:id"`
	SukukAddress   string `gorm:"column:sukuk_address"`
	DistributionId int64  `gorm:"column:distribution_id"`
	PaymentToken   string `gorm:"column:payment_token"`
	Amount         string `gorm:"column:amount"`
	Timestamp      int64  `gorm:"column:timestamp"`
	BlockNumber    int64  `gorm:"column:block_number"`
	TxHash         string `gorm:"column:tx_hash"`
}

type IndexerYieldClaimed struct {
	ID             string `gorm:"column:id"`
	User           string `gorm:"column:user"`
	SukukAddress   string `gorm:"column:sukuk_address"`
	DistributionId int64  `gorm:"column:distribution_id"`
	Amount         string `gorm:"column:amount"`
	Timestamp      int64  `gorm:"column:timestamp"`
	BlockNumber    int64  `gorm:"column:block_number"`
	TxHash         string `gorm:"column:tx_hash"`
}

type IndexerSnapshotTaken struct {
	ID            string `gorm:"column:id"`
	SukukAddress  string `gorm:"column:sukuk_address"`
	SnapshotId    int64  `gorm:"column:snapshot_id"`
	TotalSupply   string `gorm:"column:total_supply"`
	HolderCount   int64  `gorm:"column:holder_count"`
	EligibleCount int64  `gorm:"column:eligible_count"`
	Timestamp     int64  `gorm:"column:timestamp"`
	BlockNumber   int64  `gorm:"column:block_number"`
	TxHash        string `gorm:"column:tx_hash"`
}

type IndexerHolderUpdated struct {
	ID           string `gorm:"column:id"`
	SukukAddress string `gorm:"column:sukuk_address"`
	Holder       string `gorm:"column:holder"`
	Balance      string `gorm:"column:new_balance"`
	Timestamp    int64  `gorm:"column:timestamp"`
	BlockNumber  int64  `gorm:"column:block_number"`
	TxHash       string `gorm:"column:tx_hash"`
}

// Portfolio and yield calculation methods

// GetUserPortfolio calculates user's portfolio with holdings and claimable yields
func (s *IndexerQueryService) GetUserPortfolio(userAddress string) (*UserPortfolio, error) {
	if s.indexerDB == nil {
		if err := s.ConnectToIndexer(); err != nil {
			return nil, err
		}
	}

	portfolio := &UserPortfolio{
		Address:  userAddress,
		Holdings: []SukukHolding{},
	}

	// Get all sukuk addresses owned by user
	sukukAddresses, err := s.GetSukukOwnedByAddress(userAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get owned sukuk: %w", err)
	}

	// For each sukuk, calculate holdings and claimable yields
	for _, sukukAddr := range sukukAddresses {
		holding, err := s.GetSukukHolding(userAddress, sukukAddr)
		if err != nil {
			// Log error but continue with other sukuk
			continue
		}
		if holding != nil {
			portfolio.Holdings = append(portfolio.Holdings, *holding)
		}
	}

	return portfolio, nil
}

// GetSukukHolding calculates user's holding and claimable yield for a specific sukuk
func (s *IndexerQueryService) GetSukukHolding(userAddress, sukukAddress string) (*SukukHolding, error) {
	// Get current balance from holder_update table
	balance, err := s.GetCurrentBalance(userAddress, sukukAddress)
	if err != nil {
		return nil, err
	}

	if balance == "0" {
		return nil, nil // User doesn't hold this sukuk anymore
	}

	// Get claimable yield
	claimableYield, err := s.GetClaimableYield(userAddress, sukukAddress)
	if err != nil {
		return nil, err
	}

	holding := &SukukHolding{
		SukukAddress:   sukukAddress,
		Balance:        balance,
		ClaimableYield: claimableYield,
	}

	return holding, nil
}

// GetCurrentBalance gets user's current balance for a sukuk from holder_update table
func (s *IndexerQueryService) GetCurrentBalance(userAddress, sukukAddress string) (string, error) {
	holderTable, err := s.tableService.GetLatestTableForEvent("holder_update")
	if err != nil {
		return "0", fmt.Errorf("failed to find holder_update table: %w", err)
	}

	var holder IndexerHolderUpdated
	err = s.indexerDB.Table(holderTable).
		Where("holder = ? AND sukuk_address = ?", userAddress, sukukAddress).
		Order("timestamp DESC").
		First(&holder).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "0", nil
		}
		return "0", fmt.Errorf("failed to query current balance: %w", err)
	}

	return holder.Balance, nil
}

// GetClaimableYield calculates claimable yield by comparing distributed vs claimed
func (s *IndexerQueryService) GetClaimableYield(userAddress, sukukAddress string) (string, error) {
	mathUtil := utils.GlobalTokenMath
	
	// Get total yield distributed for this sukuk
	totalDistributed, err := s.GetTotalYieldDistributed(sukukAddress)
	if err != nil {
		return "0", err
	}

	// Get total yield claimed by user for this sukuk  
	totalClaimed, err := s.GetTotalYieldClaimed(userAddress, sukukAddress)
	if err != nil {
		return "0", err
	}

	// Get user's share percentage based on current holdings
	// This is simplified - ideally should check balance at each distribution snapshot
	sharePercentage, err := s.GetUserSharePercentage(userAddress, sukukAddress)
	if err != nil {
		return "0", err
	}

	// Calculate user's entitled yield = totalDistributed * sharePercentage
	entitledYield, err := mathUtil.MultiplyTokenAmount(totalDistributed, sharePercentage)
	if err != nil {
		return "0", fmt.Errorf("failed to calculate entitled yield: %w", err)
	}

	// Calculate claimable = entitledYield - totalClaimed
	claimableYield, err := mathUtil.SubtractTokenAmounts(entitledYield, totalClaimed)
	if err != nil {
		return "0", fmt.Errorf("failed to calculate claimable yield: %w", err)
	}

	return claimableYield, nil
}

// GetTotalYieldDistributed gets total yield distributed for a sukuk
func (s *IndexerQueryService) GetTotalYieldDistributed(sukukAddress string) (string, error) {
	yieldTable, err := s.tableService.GetLatestTableForEvent("yield_distributed")
	if err != nil {
		return "0", fmt.Errorf("failed to find yield_distributed table: %w", err)
	}

	var yields []IndexerYieldDistributed
	err = s.indexerDB.Table(yieldTable).
		Where("sukuk_address = ?", sukukAddress).
		Find(&yields).Error

	if err != nil {
		return "0", fmt.Errorf("failed to query yield distributions: %w", err)
	}

	// Sum all distributed amounts using proper BigInt math
	mathUtil := utils.GlobalTokenMath
	total := "0"
	
	for _, y := range yields {
		newTotal, err := mathUtil.AddTokenAmounts(total, y.Amount)
		if err != nil {
			return "0", fmt.Errorf("failed to sum yield amounts: %w", err)
		}
		total = newTotal
	}

	return total, nil
}

// GetTotalYieldClaimed gets total yield claimed by user for a sukuk
func (s *IndexerQueryService) GetTotalYieldClaimed(userAddress, sukukAddress string) (string, error) {
	claimedTable, err := s.tableService.GetLatestTableForEvent("yield_claimed")
	if err != nil {
		return "0", fmt.Errorf("failed to find yield_claimed table: %w", err)
	}

	var claims []IndexerYieldClaimed
	err = s.indexerDB.Table(claimedTable).
		Where("user = ? AND sukuk_address = ?", userAddress, sukukAddress).
		Find(&claims).Error

	if err != nil {
		return "0", fmt.Errorf("failed to query yield claims: %w", err)
	}

	// Sum all claimed amounts using proper BigInt math
	mathUtil := utils.GlobalTokenMath
	total := "0"
	
	for _, c := range claims {
		newTotal, err := mathUtil.AddTokenAmounts(total, c.Amount)
		if err != nil {
			return "0", fmt.Errorf("failed to sum claimed amounts: %w", err)
		}
		total = newTotal
	}

	return total, nil
}

// GetUserSharePercentage calculates user's ownership percentage of a sukuk
func (s *IndexerQueryService) GetUserSharePercentage(userAddress, sukukAddress string) (float64, error) {
	mathUtil := utils.GlobalTokenMath
	
	// Get user's current balance
	userBalance, err := s.GetCurrentBalance(userAddress, sukukAddress)
	if err != nil {
		return 0, err
	}

	// If user has no balance, share is 0%
	if mathUtil.IsZero(userBalance) {
		return 0.0, nil
	}

	// Get total supply from latest redemption request (which includes totalSupply)
	// or snapshot table which also tracks totalSupply
	totalSupply, err := s.getTotalSupplyFromSnapshot(sukukAddress)
	if err != nil {
		// Fallback: try to get from redemption events
		totalSupply, err = s.getTotalSupplyFromRedemption(sukukAddress)
		if err != nil {
			return 0, fmt.Errorf("failed to get total supply: %w", err)
		}
	}

	// Calculate percentage: userBalance / totalSupply
	percentage, err := mathUtil.CalculatePercentage(userBalance, totalSupply)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate percentage: %w", err)
	}

	return percentage, nil
}

// GetYieldDistributions gets yield distribution events for a sukuk
func (s *IndexerQueryService) GetYieldDistributions(sukukAddress string, limit int) ([]IndexerYieldDistributed, error) {
	yieldTable, err := s.tableService.GetLatestTableForEvent("yield_distributed")
	if err != nil {
		return nil, fmt.Errorf("failed to find yield_distributed table: %w", err)
	}

	var yields []IndexerYieldDistributed
	query := s.indexerDB.Table(yieldTable).
		Order("timestamp DESC")

	if sukukAddress != "" {
		query = query.Where("sukuk_address = ?", sukukAddress)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	err = query.Find(&yields).Error
	return yields, err
}

// GetYieldClaims gets yield claim events for a user/sukuk
func (s *IndexerQueryService) GetYieldClaims(userAddress, sukukAddress string, limit int) ([]IndexerYieldClaimed, error) {
	claimedTable, err := s.tableService.GetLatestTableForEvent("yield_claimed")
	if err != nil {
		return nil, fmt.Errorf("failed to find yield_claimed table: %w", err)
	}

	var claims []IndexerYieldClaimed
	query := s.indexerDB.Table(claimedTable).
		Order("timestamp DESC")

	if userAddress != "" {
		query = query.Where("user = ?", userAddress)
	}

	if sukukAddress != "" {
		query = query.Where("sukuk_address = ?", sukukAddress)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	err = query.Find(&claims).Error
	return claims, err
}

// getTotalSupplyFromSnapshot gets total supply from snapshot table
func (s *IndexerQueryService) getTotalSupplyFromSnapshot(sukukAddress string) (string, error) {
	snapshotTable, err := s.tableService.GetLatestTableForEvent("snapshot")
	if err != nil {
		return "0", fmt.Errorf("failed to find snapshot table: %w", err)
	}

	var snapshot struct {
		TotalSupply string `gorm:"column:total_supply"`
	}

	err = s.indexerDB.Table(snapshotTable).
		Where("sukuk_address = ?", sukukAddress).
		Order("timestamp DESC").
		First(&snapshot).Error

	if err != nil {
		return "0", fmt.Errorf("failed to get total supply from snapshot: %w", err)
	}

	return snapshot.TotalSupply, nil
}

// getTotalSupplyFromRedemption gets total supply from latest redemption request
func (s *IndexerQueryService) getTotalSupplyFromRedemption(sukukAddress string) (string, error) {
	redemptionTable, err := s.tableService.GetLatestTableForEvent("redemption_request")
	if err != nil {
		return "0", fmt.Errorf("failed to find redemption_request table: %w", err)
	}

	var redemption struct {
		TotalSupply string `gorm:"column:total_supply"`
	}

	err = s.indexerDB.Table(redemptionTable).
		Where("sukuk_address = ?", sukukAddress).
		Order("timestamp DESC").
		First(&redemption).Error

	if err != nil {
		return "0", fmt.Errorf("failed to get total supply from redemption: %w", err)
	}

	return redemption.TotalSupply, nil
}

// GetUserTransactionHistory gets all transactions for a user efficiently with database-level filtering and sorting
func (s *IndexerQueryService) GetUserTransactionHistory(userAddress string, limit int) ([]models.TransactionEvent, error) {
	if s.indexerDB == nil {
		if err := s.ConnectToIndexer(); err != nil {
			return nil, err
		}
	}

	allTransactions := make([]models.TransactionEvent, 0)

	// Get purchases with database filtering
	purchaseTable, err := s.tableService.GetLatestTableForEvent("sukuk_purchase")
	if err == nil {
		var purchases []IndexerSukukPurchase
		err = s.indexerDB.Table(purchaseTable).
			Where("buyer = ?", userAddress).
			Order("timestamp DESC").
			Limit(limit).
			Find(&purchases).Error

		if err == nil {
			for _, p := range purchases {
				allTransactions = append(allTransactions, models.TransactionEvent{
					Type:         "purchase",
					SukukAddress: p.SukukAddress,
					Amount:       p.Amount,
					TxHash:       p.TxHash,
					Timestamp:    time.Unix(p.Timestamp, 0),
					BlockNumber:  p.BlockNumber,
					Status:       "confirmed",
					Details: map[string]interface{}{
						"payment_token": p.PaymentToken,
						"buyer":         p.Buyer,
					},
				})
			}
		}
	}

	// Get redemption requests with database filtering
	redemptionTable, err := s.tableService.GetLatestTableForEvent("redemption_request")
	if err == nil {
		var redemptions []IndexerRedemptionRequest
		err = s.indexerDB.Table(redemptionTable).
			Where("user = ?", userAddress).
			Order("timestamp DESC").
			Limit(limit).
			Find(&redemptions).Error

		if err == nil {
			for _, r := range redemptions {
				allTransactions = append(allTransactions, models.TransactionEvent{
					Type:         "redemption_request",
					SukukAddress: r.SukukAddress,
					Amount:       r.Amount,
					TxHash:       r.TxHash,
					Timestamp:    time.Unix(r.Timestamp, 0),
					BlockNumber:  r.BlockNumber,
					Status:       "confirmed",
					Details: map[string]interface{}{
						"payment_token": r.PaymentToken,
						"user":          r.User,
					},
				})
			}
		}
	}

	// Get yield claims with database filtering
	yieldTable, err := s.tableService.GetLatestTableForEvent("yield_claimed")
	if err == nil {
		var claims []IndexerYieldClaimed
		err = s.indexerDB.Table(yieldTable).
			Where("user = ?", userAddress).
			Order("timestamp DESC").
			Limit(limit).
			Find(&claims).Error

		if err == nil {
			for _, y := range claims {
				allTransactions = append(allTransactions, models.TransactionEvent{
					Type:         "yield_claim",
					SukukAddress: y.SukukAddress,
					Amount:       y.Amount,
					TxHash:       y.TxHash,
					Timestamp:    time.Unix(y.Timestamp, 0),
					BlockNumber:  y.BlockNumber,
					Status:       "confirmed",
					Details: map[string]interface{}{
						"user": y.User,
					},
				})
			}
		}
	}

	// Sort by timestamp descending using Go's sort package (more efficient than bubble sort)
	sort.Slice(allTransactions, func(i, j int) bool {
		return allTransactions[i].Timestamp.After(allTransactions[j].Timestamp)
	})

	// Apply final limit
	if len(allTransactions) > limit {
		allTransactions = allTransactions[:limit]
	}

	return allTransactions, nil
}

// GetAvailableTables returns all available indexer tables with their event types
func (s *IndexerQueryService) GetAvailableTables() (map[string]string, error) {
	return s.tableService.GetAllLatestTables()
}

// GetAvailableDistributions gets yield distributions for a sukuk with claim information for a specific user
func (s *IndexerQueryService) GetAvailableDistributions(userAddress, sukukAddress string) ([]models.SukukYieldDistribution, error) {
	// Always return an empty slice if there are any errors - don't fail the entire owned-sukuk response
	emptyResult := []models.SukukYieldDistribution{}

	if s.indexerDB == nil {
		if err := s.ConnectToIndexer(); err != nil {
			return emptyResult, nil // Return empty instead of error
		}
	}

	// Get latest table names using dynamic discovery
	distributionTable, err := s.tableService.GetLatestTableForEvent("yield_distribution")
	if err != nil {
		return emptyResult, nil // Return empty instead of error
	}

	claimTable, err := s.tableService.GetLatestTableForEvent("yield_claim")
	if err != nil {
		return emptyResult, nil // Return empty instead of error
	}

	// Get all yield distributions for this sukuk
	var distributions []IndexerYieldDistributed
	err = s.indexerDB.Table(distributionTable).
		Where("sukuk_address = ?", sukukAddress).
		Order("distribution_id ASC").
		Find(&distributions).Error
	if err != nil {
		return emptyResult, nil // Return empty instead of error
	}

	// If no distributions found, return empty result
	if len(distributions) == 0 {
		return emptyResult, nil
	}

	// Get all yield claims by this user for this sukuk
	var claims []IndexerYieldClaimed
	err = s.indexerDB.Table(claimTable).
		Where("user = ? AND sukuk_address = ?", userAddress, sukukAddress).
		Find(&claims).Error
	if err != nil {
		return emptyResult, nil // Return empty instead of error
	}

	// Create a map of claimed amounts by distribution ID
	claimedMap := make(map[int64]string)
	for _, claim := range claims {
		if existingAmount, exists := claimedMap[claim.DistributionId]; exists {
			// Sum up multiple claims for the same distribution
			mathUtil := utils.GlobalTokenMath
			newTotal, err := mathUtil.AddTokenAmounts(existingAmount, claim.Amount)
			if err != nil {
				continue // Skip on error, log if needed
			}
			claimedMap[claim.DistributionId] = newTotal
		} else {
			claimedMap[claim.DistributionId] = claim.Amount
		}
	}

	// Get user's current balance to calculate claimable amount
	userBalance, err := s.GetCurrentBalance(userAddress, sukukAddress)
	if err != nil {
		userBalance = "0" // Default to 0 if error
	}

	// Get total supply to calculate user's share
	totalSupply, err := s.getTotalSupplyFromSnapshot(sukukAddress)
	if err != nil {
		// Fallback to redemption data
		totalSupply, err = s.getTotalSupplyFromRedemption(sukukAddress)
		if err != nil {
			totalSupply = "0" // Default to 0 if error
		}
	}

	// Build result
	result := make([]models.SukukYieldDistribution, len(distributions))
	mathUtil := utils.GlobalTokenMath
	
	for i, dist := range distributions {
		claimedAmount := "0"
		if amount, exists := claimedMap[dist.DistributionId]; exists {
			claimedAmount = amount
		}

		// Calculate user's claimable amount based on their share
		userClaimableAmount := "0"
		claimable := false
		
		if !mathUtil.IsZero(userBalance) && !mathUtil.IsZero(totalSupply) {
			// Calculate user's percentage: userBalance / totalSupply
			percentage, err := mathUtil.CalculatePercentage(userBalance, totalSupply)
			if err == nil {
				// Calculate user's entitled amount: distribution.Amount * percentage
				entitledAmount, err := mathUtil.MultiplyTokenAmount(dist.Amount, percentage)
				if err == nil {
					// Calculate claimable: entitledAmount - claimedAmount
					userClaimableAmount, err = mathUtil.SubtractTokenAmounts(entitledAmount, claimedAmount)
					if err == nil && !mathUtil.IsZero(userClaimableAmount) {
						claimable = true
					}
				}
			}
		}

		result[i] = models.SukukYieldDistribution{
			DistributionId:      dist.DistributionId,
			Amount:              dist.Amount,
			PaymentToken:        dist.PaymentToken,
			Claimable:           claimable,
			ClaimedAmount:       claimedAmount,
			UserClaimableAmount: userClaimableAmount,
		}
	}

	return result, nil
}

// Portfolio calculation result structures
type UserPortfolio struct {
	Address  string         `json:"address"`
	Holdings []SukukHolding `json:"holdings"`
}

type SukukHolding struct {
	SukukAddress   string `json:"sukuk_address"`
	Balance        string `json:"balance"`
	ClaimableYield string `json:"claimable_yield"`
}