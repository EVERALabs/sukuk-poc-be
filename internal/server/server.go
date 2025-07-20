package server

import (
	"fmt"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kadzu/sukuk-poc-be/internal/config"
	"github.com/kadzu/sukuk-poc-be/internal/handlers"
	"github.com/kadzu/sukuk-poc-be/internal/logger"
	"github.com/kadzu/sukuk-poc-be/internal/middleware"
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
		AllowOrigins:     cfg.API.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-API-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
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
	
	// Health check endpoint (no auth required)
	s.router.GET("/health", handlers.Health)

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
			sukuk.GET("", handlers.ListSukukSeries)
			sukuk.GET("/:id", handlers.GetSukukSeries)
			sukuk.GET("/:id/metrics", handlers.GetSukukMetrics)
			sukuk.GET("/:id/holders", handlers.GetSukukHolders)
		}

		// Portfolio endpoints
		v1.GET("/portfolio/:address", handlers.GetPortfolio)
		v1.GET("/portfolio/:address/investments", handlers.GetInvestmentHistory)
		v1.GET("/portfolio/:address/yields", handlers.GetYieldHistory)
		v1.GET("/portfolio/:address/yields/pending", handlers.GetPendingYields)
		v1.GET("/portfolio/:address/redemptions", handlers.GetRedemptionHistory)
		
		// Investment endpoints (read-only from blockchain events)
		investments := v1.Group("/investments")
		{
			investments.GET("", handlers.GetInvestments)
			investments.GET("/:id", handlers.GetInvestment)
			investments.GET("/investor/:address", handlers.GetInvestmentsByInvestor)
			investments.GET("/sukuk/:sukukId", handlers.GetInvestmentsBySukuk)
		}

		// Yield Claims endpoints
		yields := v1.Group("/yield-claims")
		{
			yields.GET("", handlers.GetYieldClaims)
			yields.GET("/:id", handlers.GetYieldClaim)
			yields.GET("/investor/:address", handlers.GetYieldClaimsByInvestor)
			yields.GET("/sukuk/:sukukId", handlers.GetYieldClaimsBySukuk)
		}

		// Redemption endpoints
		redemptions := v1.Group("/redemptions")
		{
			redemptions.GET("", handlers.GetRedemptions)
			redemptions.GET("/:id", handlers.GetRedemption)
			redemptions.GET("/investor/:address", handlers.GetRedemptionsByInvestor)
			redemptions.GET("/sukuk/:sukukId", handlers.GetRedemptionsBySukuk)
			redemptions.POST("", handlers.CreateRedemption)
			redemptions.PUT("/:id/approve", handlers.ApproveRedemption)
			redemptions.PUT("/:id/reject", handlers.RejectRedemption)
		}
		
		// Analytics endpoints
		v1.GET("/analytics/overview", handlers.GetPlatformStats)
		v1.GET("/analytics/vault/:seriesId", handlers.GetVaultBalance)
		
		// Event endpoints
		v1.GET("/events/:txHash", handlers.GetEventByTxHash)
		
		// Protected endpoints (require API key)
		protected := v1.Group("/admin")
		protected.Use(middleware.APIKeyAuth(s.cfg.API.APIKey))
		{
			// Company management
			protected.POST("/companies", handlers.CreateCompany)
			protected.PUT("/companies/:id", handlers.UpdateCompany)
			protected.POST("/companies/:id/upload-logo", handlers.UploadCompanyLogo)
			
			// Sukuk management  
			protected.POST("/sukuks", handlers.CreateSukukSeries)
			protected.PUT("/sukuks/:id", handlers.UpdateSukukSeries)
			protected.POST("/sukuks/:id/upload-prospectus", handlers.UploadProspectus)
			
			// Webhook for indexer
			protected.POST("/events/webhook", handlers.ProcessEventWebhook)
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