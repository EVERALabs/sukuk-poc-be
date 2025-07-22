package server

import (
	"fmt"
	"time"

	"sukuk-be/internal/config"
	"sukuk-be/internal/handlers"
	"sukuk-be/internal/logger"
	"sukuk-be/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	cfg    *config.Config
	router *gin.Engine
}

func New(cfg *config.Config) *Server {
	// Set gin mode based on environment
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(middleware.RequestLogger())
	router.Use(middleware.ErrorLogger())
	router.Use(gin.Recovery())

	// CORS middleware
	corsConfig := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-API-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	
	// Check if we should allow all origins
	if len(cfg.API.AllowedOrigins) == 1 && cfg.API.AllowedOrigins[0] == "*" {
		corsConfig.AllowAllOrigins = true
		corsConfig.AllowCredentials = false // Cannot use credentials with AllowAllOrigins
	} else {
		corsConfig.AllowOrigins = cfg.API.AllowedOrigins
	}
	
	router.Use(cors.New(corsConfig))

	return &Server{
		cfg:    cfg,
		router: router,
	}
}

func (s *Server) setupRoutes() {
	// Static file serving for uploads
	s.router.Static("/uploads", "./uploads")

	// Swagger documentation
	s.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check endpoint (no auth required)
	s.router.GET("/health", handlers.GetHealthStatus)

	// API v1 group with middleware
	v1 := s.router.Group("/api/v1")
	v1.Use(middleware.RateLimit(s.cfg.API.RateLimitPerMin))
	{
		// Company endpoints
		companies := v1.Group("/companies")
		{
			companies.GET("", handlers.ListCompanies)
			companies.GET("/:id", handlers.GetCompany)
			companies.GET("/:id/sukuks", handlers.GetCompanySukuks)
		}

		// Sukuk Series endpoints
		sukuk := v1.Group("/sukuks")
		{
			sukuk.GET("", handlers.ListSukuk)
			sukuk.GET("/:id", handlers.GetSukuk)
			sukuk.GET("/:id/metrics", handlers.GetSukukMetrics)
			// TODO: Implement GetSukukMetricsWithBlockchain handler
			// sukuk.GET("/:id/blockchain-metrics", handlers.GetSukukMetricsWithBlockchain)
			sukuk.GET("/:id/holders", handlers.GetSukukHolders)
		}

		// Portfolio endpoints
		// TODO: Implement GetPortfolio handler
		// v1.GET("/portfolio/:address", handlers.GetPortfolio)
		// TODO: Implement GetPortfolioWithBlockchainData handler
		// v1.GET("/portfolio/:address/blockchain", handlers.GetPortfolioWithBlockchainData)
		// Use existing investment portfolio handler
		v1.GET("/portfolio/:address/investments", handlers.GetInvestmentPortfolio)
		// TODO: Implement GetYieldHistory handler
		// v1.GET("/portfolio/:address/yields", handlers.GetYieldHistory)
		v1.GET("/portfolio/:address/yields/pending", handlers.GetPendingYields)
		// TODO: Implement GetRedemptionHistory handler
		// v1.GET("/portfolio/:address/redemptions", handlers.GetRedemptionHistory)

		// Investment endpoints (read-only from blockchain events)
		investments := v1.Group("/investments")
		{
			investments.GET("", handlers.ListInvestments)
			// TODO: Implement GetInvestment handler
			// investments.GET("/:id", handlers.GetInvestment)
			// TODO: Implement GetInvestmentWithBlockchainData handler
			// investments.GET("/:id/blockchain", handlers.GetInvestmentWithBlockchainData)
			investments.GET("/investor/:address", handlers.GetInvestmentsByInvestor)
			// TODO: Implement GetInvestmentsBySukuk handler
			// investments.GET("/sukuk/:sukukId", handlers.GetInvestmentsBySukuk)
		}

		// Yield Claims endpoints
		yields := v1.Group("/yield-claims")
		{
			yields.GET("", handlers.ListYields)
			// TODO: Implement GetYieldClaim handler
			// yields.GET("/:id", handlers.GetYieldClaim)
			yields.GET("/investor/:address", handlers.GetYieldsByInvestor)
			yields.GET("/sukuk/:sukukId", handlers.GetYieldsBySukuk)
		}

		// Redemption endpoints
		redemptions := v1.Group("/redemptions")
		{
			redemptions.GET("", handlers.ListRedemptions)
			// TODO: Implement GetRedemption handler
			// redemptions.GET("/:id", handlers.GetRedemption)
			redemptions.GET("/investor/:address", handlers.GetRedemptionsByInvestor)
			redemptions.GET("/sukuk/:sukukId", handlers.GetRedemptionsBySukuk)
			// TODO: Implement CreateRedemption handler
			// redemptions.POST("", handlers.CreateRedemption)
			// TODO: Implement ApproveRedemption handler
			// redemptions.PUT("/:id/approve", handlers.ApproveRedemption)
			// TODO: Implement RejectRedemption handler
			// redemptions.PUT("/:id/reject", handlers.RejectRedemption)
		}

		// Analytics endpoints
		// TODO: Implement GetPlatformStats handler
		// v1.GET("/analytics/overview", handlers.GetPlatformStats)
		// TODO: Implement GetVaultBalance handler
		// v1.GET("/analytics/vault/:seriesId", handlers.GetVaultBalance)

		// Blockchain data endpoints
		// TODO: Implement blockchain endpoints when handlers are ready
		// blockchain := v1.Group("/blockchain")
		// {
		//		blockchain.GET("/events/:txHash", handlers.GetBlockchainEventsByTxHash)
		// }

		// Protected endpoints (require API key)
		protected := v1.Group("/admin")
		protected.Use(middleware.APIKeyAuth(s.cfg.API.APIKey))
		{
			// Company management
			protected.POST("/companies", handlers.CreateCompany)
			protected.PUT("/companies/:id", handlers.UpdateCompany)
			protected.POST("/companies/:id/upload-logo", handlers.UploadCompanyLogo)

			// Sukuk management
			protected.POST("/sukuks", handlers.CreateSukuk)
			protected.PUT("/sukuks/:id", handlers.UpdateSukuk)
			protected.POST("/sukuks/:id/upload-prospectus", handlers.UploadProspectus)
		}
	}
}

func (s *Server) Start() error {
	// Setup routes
	s.setupRoutes()

	// Start server
	addr := fmt.Sprintf(":%d", s.cfg.App.Port)
	logger.WithField("address", addr).Info("Server listening and serving HTTP")

	return s.router.Run(addr)
}
