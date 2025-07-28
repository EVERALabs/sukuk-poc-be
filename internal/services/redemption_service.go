package services

import (
	"fmt"
	"time"

	"sukuk-be/internal/database"
	"sukuk-be/internal/models"
	"sukuk-be/internal/utils"
)

type RedemptionService struct {
	indexerService *IndexerQueryService
	mathUtil       *utils.TokenMath
}

func NewRedemptionService() *RedemptionService {
	return &RedemptionService{
		indexerService: NewIndexerQueryService(),
		mathUtil:       utils.GlobalTokenMath,
	}
}

// GetAllRedemptions returns all redemptions with their approval status
func (s *RedemptionService) GetAllRedemptions(limit int, offset int) (*models.RedemptionListResponse, error) {
	if err := s.indexerService.ConnectToIndexer(); err != nil {
		return nil, err
	}

	// Get redemption requests
	requests, err := s.getRedemptionRequests(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get redemption requests: %w", err)
	}

	// Get redemption approvals
	approvals, err := s.getRedemptionApprovals()
	if err != nil {
		return nil, fmt.Errorf("failed to get redemption approvals: %w", err)
	}

	// Merge and create comprehensive redemption list
	redemptions := s.mergeRedemptionsWithApprovals(requests, approvals)

	// Add metadata for each redemption
	for i := range redemptions {
		var sukukMetadata models.SukukMetadata
		if err := database.GetDB().Where("contract_address = ?", redemptions[i].SukukAddress).First(&sukukMetadata).Error; err == nil {
			redemptions[i].Metadata = &sukukMetadata
		}
		
		// Determine if can be approved (not already approved)
		redemptions[i].CanApprove = redemptions[i].Status == models.RedemptionStatusRequested
		redemptions[i].RequiresManagerAuth = true
	}

	// Calculate status counts
	statusCounts := make(map[string]int)
	for _, r := range redemptions {
		statusCounts[string(r.Status)]++
	}

	return &models.RedemptionListResponse{
		TotalCount:   len(redemptions),
		Redemptions:  redemptions,
		StatusCounts: statusCounts,
	}, nil
}

// GetRedemptionsByUser returns redemptions for a specific user
func (s *RedemptionService) GetRedemptionsByUser(userAddress string) (*models.RedemptionListResponse, error) {
	if err := s.indexerService.ConnectToIndexer(); err != nil {
		return nil, err
	}

	// Get user's redemption requests
	requests, err := s.getUserRedemptionRequests(userAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get user redemption requests: %w", err)
	}

	// Get user's redemption approvals
	approvals, err := s.getUserRedemptionApprovals(userAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get user redemption approvals: %w", err)
	}

	// Merge and create comprehensive redemption list
	redemptions := s.mergeRedemptionsWithApprovals(requests, approvals)

	return &models.RedemptionListResponse{
		TotalCount:  len(redemptions),
		Redemptions: redemptions,
	}, nil
}

// GetRedemptionsBySukuk returns redemptions for a specific sukuk
func (s *RedemptionService) GetRedemptionsBySukuk(sukukAddress string) (*models.RedemptionListResponse, error) {
	if err := s.indexerService.ConnectToIndexer(); err != nil {
		return nil, err
	}

	// Get sukuk's redemption requests
	requests, err := s.getSukukRedemptionRequests(sukukAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get sukuk redemption requests: %w", err)
	}

	// Get sukuk's redemption approvals
	approvals, err := s.getSukukRedemptionApprovals(sukukAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get sukuk redemption approvals: %w", err)
	}

	// Merge and create comprehensive redemption list
	redemptions := s.mergeRedemptionsWithApprovals(requests, approvals)

	return &models.RedemptionListResponse{
		TotalCount:  len(redemptions),
		Redemptions: redemptions,
	}, nil
}

// GetRedemptionStats returns overall redemption statistics
func (s *RedemptionService) GetRedemptionStats() (*models.RedemptionStatsResponse, error) {
	allRedemptions, err := s.GetAllRedemptions(1000, 0) // Get a large set for stats
	if err != nil {
		return nil, err
	}

	stats := &models.RedemptionStatsResponse{
		TotalRequests:   len(allRedemptions.Redemptions),
		BySukuk:         make(map[string]models.RedemptionSukukStats),
	}

	var totalRequestedAmounts, totalApprovedAmounts []string
	sukukStats := make(map[string]*models.RedemptionSukukStats)

	for _, r := range allRedemptions.Redemptions {
		// Overall stats
		if r.Status == models.RedemptionStatusRequested {
			stats.PendingRequests++
		} else if r.Status == models.RedemptionStatusApproved {
			stats.ApprovedRequests++
		}

		totalRequestedAmounts = append(totalRequestedAmounts, r.Amount)
		if r.ApprovedAmount != nil {
			totalApprovedAmounts = append(totalApprovedAmounts, *r.ApprovedAmount)
		}

		// Per-sukuk stats
		if _, exists := sukukStats[r.SukukAddress]; !exists {
			sukukCode := ""
			if r.Metadata != nil {
				sukukCode = r.Metadata.SukukCode
			}
			sukukStats[r.SukukAddress] = &models.RedemptionSukukStats{
				SukukAddress: r.SukukAddress,
				SukukCode:    sukukCode,
			}
		}
		
		sukuk := sukukStats[r.SukukAddress]
		sukuk.RequestCount++
		
		// Sum amounts using proper BigInt math
		if sukuk.RequestedAmount == "" {
			sukuk.RequestedAmount = r.Amount
		} else {
			if sum, err := s.mathUtil.AddTokenAmounts(sukuk.RequestedAmount, r.Amount); err == nil {
				sukuk.RequestedAmount = sum
			}
		}
		
		if r.ApprovedAmount != nil {
			if sukuk.ApprovedAmount == "" {
				sukuk.ApprovedAmount = *r.ApprovedAmount
			} else {
				if sum, err := s.mathUtil.AddTokenAmounts(sukuk.ApprovedAmount, *r.ApprovedAmount); err == nil {
					sukuk.ApprovedAmount = sum
				}
			}
		}
	}

	// Calculate totals
	if totalRequested, err := s.mathUtil.SumTokenAmounts(totalRequestedAmounts); err == nil {
		stats.TotalRequestedAmount = totalRequested
	}
	if totalApproved, err := s.mathUtil.SumTokenAmounts(totalApprovedAmounts); err == nil {
		stats.TotalApprovedAmount = totalApproved
	}

	// Convert map to response format
	for addr, stat := range sukukStats {
		stats.BySukuk[addr] = *stat
	}

	return stats, nil
}

// Private helper methods

func (s *RedemptionService) getRedemptionRequests(limit, offset int) ([]IndexerRedemptionRequest, error) {
	tableService := NewIndexerTableService()
	if err := tableService.ConnectToIndexer(); err != nil {
		return nil, err
	}

	requestTable, err := tableService.GetLatestTableForEvent("redemption_request")
	if err != nil {
		return nil, fmt.Errorf("failed to find redemption_request table: %w", err)
	}

	var requests []IndexerRedemptionRequest
	query := s.indexerService.indexerDB.Table(requestTable).
		Order("timestamp DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err = query.Find(&requests).Error
	return requests, err
}

func (s *RedemptionService) getRedemptionApprovals() ([]IndexerRedemptionApproval, error) {
	tableService := NewIndexerTableService()
	if err := tableService.ConnectToIndexer(); err != nil {
		return nil, err
	}

	approvalTable, err := tableService.GetLatestTableForEvent("redemption_approval")
	if err != nil {
		return nil, fmt.Errorf("failed to find redemption_approval table: %w", err)
	}

	var approvals []IndexerRedemptionApproval
	err = s.indexerService.indexerDB.Table(approvalTable).
		Order("timestamp DESC").
		Find(&approvals).Error
	
	return approvals, err
}

func (s *RedemptionService) getUserRedemptionRequests(userAddress string) ([]IndexerRedemptionRequest, error) {
	tableService := NewIndexerTableService()
	if err := tableService.ConnectToIndexer(); err != nil {
		return nil, err
	}

	requestTable, err := tableService.GetLatestTableForEvent("redemption_request")
	if err != nil {
		return nil, fmt.Errorf("failed to find redemption_request table: %w", err)
	}

	var requests []IndexerRedemptionRequest
	err = s.indexerService.indexerDB.Table(requestTable).
		Where("user = ?", userAddress).
		Order("timestamp DESC").
		Find(&requests).Error
	
	return requests, err
}

func (s *RedemptionService) getUserRedemptionApprovals(userAddress string) ([]IndexerRedemptionApproval, error) {
	tableService := NewIndexerTableService()
	if err := tableService.ConnectToIndexer(); err != nil {
		return nil, err
	}

	approvalTable, err := tableService.GetLatestTableForEvent("redemption_approval")
	if err != nil {
		return nil, fmt.Errorf("failed to find redemption_approval table: %w", err)
	}

	var approvals []IndexerRedemptionApproval
	err = s.indexerService.indexerDB.Table(approvalTable).
		Where("user = ?", userAddress).
		Order("timestamp DESC").
		Find(&approvals).Error
	
	return approvals, err
}

func (s *RedemptionService) getSukukRedemptionRequests(sukukAddress string) ([]IndexerRedemptionRequest, error) {
	tableService := NewIndexerTableService()
	if err := tableService.ConnectToIndexer(); err != nil {
		return nil, err
	}

	requestTable, err := tableService.GetLatestTableForEvent("redemption_request")
	if err != nil {
		return nil, fmt.Errorf("failed to find redemption_request table: %w", err)
	}

	var requests []IndexerRedemptionRequest
	err = s.indexerService.indexerDB.Table(requestTable).
		Where("sukuk_address = ?", sukukAddress).
		Order("timestamp DESC").
		Find(&requests).Error
	
	return requests, err
}

func (s *RedemptionService) getSukukRedemptionApprovals(sukukAddress string) ([]IndexerRedemptionApproval, error) {
	tableService := NewIndexerTableService()
	if err := tableService.ConnectToIndexer(); err != nil {
		return nil, err
	}

	approvalTable, err := tableService.GetLatestTableForEvent("redemption_approval")
	if err != nil {
		return nil, fmt.Errorf("failed to find redemption_approval table: %w", err)
	}

	var approvals []IndexerRedemptionApproval
	err = s.indexerService.indexerDB.Table(approvalTable).
		Where("sukuk_address = ?", sukukAddress).
		Order("timestamp DESC").
		Find(&approvals).Error
	
	return approvals, err
}

// mergeRedemptionsWithApprovals combines requests with their corresponding approvals
func (s *RedemptionService) mergeRedemptionsWithApprovals(requests []IndexerRedemptionRequest, approvals []IndexerRedemptionApproval) []models.RedemptionRequest {
	// Create a map of approvals by user+sukuk for quick lookup
	approvalMap := make(map[string]IndexerRedemptionApproval)
	for _, approval := range approvals {
		key := fmt.Sprintf("%s:%s", approval.User, approval.SukukAddress)
		approvalMap[key] = approval
	}

	var redemptions []models.RedemptionRequest
	for _, req := range requests {
		redemption := models.RedemptionRequest{
			RequestID:     req.ID,
			User:          req.User,
			SukukAddress:  req.SukukAddress,
			Amount:        req.Amount,
			PaymentToken:  req.PaymentToken,
			TotalSupply:   req.TotalSupply,
			RequestTxHash: req.TxHash,
			RequestTime:   time.Unix(req.Timestamp, 0),
			RequestBlock:  req.BlockNumber,
			Status:        models.RedemptionStatusRequested,
		}

		// Check if there's a corresponding approval
		key := fmt.Sprintf("%s:%s", req.User, req.SukukAddress)
		if approval, exists := approvalMap[key]; exists {
			redemption.Status = models.RedemptionStatusApproved
			redemption.ApprovalID = &approval.ID
			redemption.ApprovalTxHash = &approval.TxHash
			approvalTime := time.Unix(approval.Timestamp, 0)
			redemption.ApprovalTime = &approvalTime
			redemption.ApprovalBlock = &approval.BlockNumber
			redemption.ApprovedAmount = &approval.Amount
		}

		redemptions = append(redemptions, redemption)
	}

	return redemptions
}