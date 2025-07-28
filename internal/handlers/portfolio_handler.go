package handlers

import (
	"net/http"
	"strconv"
	"time"

	"sukuk-be/internal/database"
	"sukuk-be/internal/logger"
	"sukuk-be/internal/models"
	"sukuk-be/internal/services"
	"sukuk-be/internal/utils"

	"github.com/gin-gonic/gin"
)

// GetUserPortfolio returns user's complete portfolio with holdings and claimable yields
// @Summary Get user portfolio
// @Description Get complete portfolio showing all sukuk holdings with current balances and claimable yields
// @Tags portfolio
// @Accept json
// @Produce json
// @Param address path string true "User wallet address" Example("0xf57093Ea18E5CfF6E7bB3bb770Ae9C492277A5a9")
// @Success 200 {object} models.PortfolioResponse "User portfolio with holdings"
// @Failure 400 {object} map[string]string "Invalid address"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /portfolio/{address} [get]
func GetUserPortfolio(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Address is required",
		})
		return
	}

	// Initialize indexer query service
	indexerService := services.NewIndexerQueryService()

	// Get user portfolio from indexer
	portfolio, err := indexerService.GetUserPortfolio(address)
	if err != nil {
		logger.WithError(err).Error("Failed to get user portfolio")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get user portfolio",
		})
		return
	}

	// Initialize math utility and response
	mathUtil := utils.GlobalTokenMath
	// Initialize with empty arrays to avoid null in JSON
	holdings := make([]models.SukukHolding, len(portfolio.Holdings))
	if len(portfolio.Holdings) == 0 {
		holdings = []models.SukukHolding{}
	}

	response := models.PortfolioResponse{
		Address:       address,
		TotalHoldings: len(portfolio.Holdings),
		Holdings:      holdings,
		Summary: models.PortfolioSummary{
			TotalSukukCount:     len(portfolio.Holdings),
			TotalClaimableYield: "0",
			TotalYieldClaimed:   "0",
			ActiveSukukCount:    0,
			MaturedSukukCount:   0,
		},
	}

	// Track totals for summary
	var totalClaimableAmounts []string
	var totalClaimedAmounts []string

	// Get sukuk metadata for each holding and enrich response
	for i, holding := range portfolio.Holdings {
		// Convert service holding to API holding
		apiHolding := models.SukukHolding{
			SukukAddress:   holding.SukukAddress,
			Balance:        holding.Balance,
			ClaimableYield: holding.ClaimableYield,
		}

		// Get sukuk metadata from database
		var sukukMetadata models.SukukMetadata
		if err := database.GetDB().Where("contract_address = ?", holding.SukukAddress).First(&sukukMetadata).Error; err == nil {
			apiHolding.Metadata = &sukukMetadata
		}

		// Get recent yield distributions for this sukuk
		distributions, err := indexerService.GetYieldDistributions(holding.SukukAddress, 5)
		if err == nil && len(distributions) > 0 {
			apiHolding.YieldHistory = make([]models.YieldDistribution, len(distributions))
			for j, dist := range distributions {
				apiHolding.YieldHistory[j] = models.YieldDistribution{
					ID:           dist.ID,
					SukukAddress: dist.SukukAddress,
					Amount:       dist.Amount,
					Timestamp:    time.Unix(dist.Timestamp, 0),
					TxHash:       dist.TxHash,
					BlockNumber:  dist.BlockNumber,
				}
			}
		}

		// Get total yield claimed by user for this sukuk
		totalClaimed, err := indexerService.GetTotalYieldClaimed(address, holding.SukukAddress)
		if err == nil {
			apiHolding.TotalYieldClaimed = totalClaimed
			totalClaimedAmounts = append(totalClaimedAmounts, totalClaimed)
		}

		response.Holdings[i] = apiHolding

		// Update summary stats
		if mathUtil.IsPositive(holding.Balance) {
			response.Summary.ActiveSukukCount++
		}

		// Add claimable yield to totals
		if mathUtil.IsPositive(holding.ClaimableYield) {
			totalClaimableAmounts = append(totalClaimableAmounts, holding.ClaimableYield)
		}

		// TODO: Check if sukuk is matured and update MaturedSukukCount
		// This would require checking maturity timestamp from metadata against current time
	}

	// Calculate summary totals
	if totalClaimable, err := mathUtil.SumTokenAmounts(totalClaimableAmounts); err == nil {
		response.Summary.TotalClaimableYield = totalClaimable
	}
	
	if totalClaimed, err := mathUtil.SumTokenAmounts(totalClaimedAmounts); err == nil {
		response.Summary.TotalYieldClaimed = totalClaimed
	}

	c.JSON(http.StatusOK, response)
}

// GetYieldClaims returns available yield claims for a user
// @Summary Get available yield claims
// @Description Get all available yield claims across user's sukuk holdings
// @Tags portfolio
// @Accept json
// @Produce json
// @Param address path string true "User wallet address" Example("0xf57093Ea18E5CfF6E7bB3bb770Ae9C492277A5a9")
// @Success 200 {object} models.YieldClaimsResponse "Available yield claims"
// @Failure 400 {object} map[string]string "Invalid address"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /yield-claims/{address} [get]
func GetYieldClaims(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Address is required",
		})
		return
	}

	// Initialize indexer query service
	indexerService := services.NewIndexerQueryService()

	// Get sukuk addresses owned by user
	sukukAddresses, err := indexerService.GetSukukOwnedByAddress(address)
	if err != nil {
		logger.WithError(err).Error("Failed to get owned sukuk addresses")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get yield claims",
		})
		return
	}

	// Initialize math utility for amount calculations
	mathUtil := utils.GlobalTokenMath
	response := models.YieldClaimsResponse{
		Address:     address,
		TotalClaims: 0,
		Claims:      make([]models.YieldClaimDetail, 0), // Initialize as empty array
		TotalAmount: "0",
	}

	// Track claimable amounts for total calculation
	var claimableAmounts []string

	// For each sukuk, calculate claimable yield
	for _, sukukAddr := range sukukAddresses {
		// Get current balance
		balance, err := indexerService.GetCurrentBalance(address, sukukAddr)
		if err != nil || mathUtil.IsZero(balance) {
			continue // Skip if no balance or error
		}

		// Get claimable yield
		claimableAmount, err := indexerService.GetClaimableYield(address, sukukAddr)
		if err != nil {
			continue
		}

		// Get latest yield distributions
		distributions, err := indexerService.GetYieldDistributions(sukukAddr, 10)
		var lastDistribution *time.Time
		distributionCount := 0
		if err == nil {
			distributionCount = len(distributions)
			if len(distributions) > 0 {
				timestamp := time.Unix(distributions[0].Timestamp, 0)
				lastDistribution = &timestamp
			}
		}

		// Get sukuk metadata
		var sukukMetadata models.SukukMetadata
		if err := database.GetDB().Where("contract_address = ?", sukukAddr).First(&sukukMetadata).Error; err != nil {
			sukukMetadata.ContractAddress = sukukAddr // Fallback
		}

		claimDetail := models.YieldClaimDetail{
			SukukAddress:      sukukAddr,
			ClaimableAmount:   claimableAmount,
			LastDistribution:  lastDistribution,
			DistributionCount: distributionCount,
			UserBalance:       balance,
			Metadata:          &sukukMetadata,
		}

		response.Claims = append(response.Claims, claimDetail)
		response.TotalClaims++

		// Add claimable amount to total calculation
		if mathUtil.IsPositive(claimableAmount) {
			claimableAmounts = append(claimableAmounts, claimableAmount)
		}
	}

	// Calculate total claimable amount using proper BigInt math
	if totalAmount, err := mathUtil.SumTokenAmounts(claimableAmounts); err == nil {
		response.TotalAmount = totalAmount
	}

	c.JSON(http.StatusOK, response)
}

// GetTransactionHistory returns complete transaction history for a user
// @Summary Get transaction history
// @Description Get complete transaction history including purchases, redemptions, and yield claims
// @Tags transactions
// @Accept json
// @Produce json
// @Param address path string true "User wallet address" Example("0xf57093Ea18E5CfF6E7bB3bb770Ae9C492277A5a9")
// @Param limit query int false "Number of transactions to return" default(50) minimum(1) maximum(200)
// @Success 200 {object} models.TransactionHistoryResponse "Transaction history"
// @Failure 400 {object} map[string]string "Invalid address or parameters"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /transactions/{address} [get]
func GetTransactionHistory(c *gin.Context) {
	address := c.Param("address")
	if address == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Address is required",
		})
		return
	}

	// Parse limit parameter
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200 // Cap at 200 for performance
	}

	// Initialize indexer query service
	indexerService := services.NewIndexerQueryService()

	// Get all transactions efficiently with database-level filtering and sorting
	allTransactions, err := indexerService.GetUserTransactionHistory(address, limit)
	if err != nil {
		logger.WithError(err).Error("Failed to get user transaction history")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get transaction history",
		})
		return
	}

	// Ensure empty array instead of null
	if allTransactions == nil {
		allTransactions = []models.TransactionEvent{}
	}

	response := models.TransactionHistoryResponse{
		Address:      address,
		TotalCount:   len(allTransactions),
		Transactions: allTransactions,
	}

	c.JSON(http.StatusOK, response)
}

// GetYieldDistributions returns yield distribution history for a sukuk
// @Summary Get yield distributions
// @Description Get yield distribution history for a specific sukuk
// @Tags portfolio
// @Accept json
// @Produce json
// @Param sukuk_address path string true "Sukuk contract address"
// @Param limit query int false "Number of distributions to return" default(20)
// @Success 200 {object} map[string]interface{} "Yield distributions"
// @Failure 400 {object} map[string]string "Invalid parameters"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /yield-distributions/{sukuk_address} [get]
func GetYieldDistributions(c *gin.Context) {
	sukukAddress := c.Param("sukuk_address")
	if sukukAddress == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Sukuk address is required",
		})
		return
	}

	// Parse limit parameter
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	// Initialize indexer query service
	indexerService := services.NewIndexerQueryService()

	// Get yield distributions
	distributions, err := indexerService.GetYieldDistributions(sukukAddress, limit)
	if err != nil {
		logger.WithError(err).Error("Failed to get yield distributions")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to get yield distributions",
		})
		return
	}

	// Convert to API format
	apiDistributions := make([]models.YieldDistribution, len(distributions))
	for i, dist := range distributions {
		apiDistributions[i] = models.YieldDistribution{
			ID:           dist.ID,
			SukukAddress: dist.SukukAddress,
			Amount:       dist.Amount,
			Timestamp:    time.Unix(dist.Timestamp, 0),
			TxHash:       dist.TxHash,
			BlockNumber:  dist.BlockNumber,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"sukuk_address":   sukukAddress,
		"total_count":     len(apiDistributions),
		"distributions":   apiDistributions,
	})
}