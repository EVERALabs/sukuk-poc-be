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
		// Sukuk Metadata endpoints (core functionality)
		sukukMetadata := v1.Group("/sukuk-metadata")
		{
			sukukMetadata.GET("", handlers.ListSukukMetadata)
			sukukMetadata.GET("/:id", handlers.GetSukukMetadata)
			sukukMetadata.POST("", handlers.CreateSukukMetadata)
			sukukMetadata.PUT("/:id", handlers.UpdateSukukMetadata)
			sukukMetadata.PUT("/:id/ready", handlers.MarkSukukMetadataReady)
			sukukMetadata.PUT("/:id/unready", handlers.MarkSukukMetadataUnready)
			sukukMetadata.POST("/sync", handlers.TriggerSukukMetadataSync)
			sukukMetadata.GET("/tables", handlers.ListSukukCreationTables)
		}

		// Transaction History endpoints (both old and new formats)
		v1.GET("/transaction-history/:address", handlers.GetRiwayatByAddress)
		v1.GET("/transactions/:address", handlers.GetTransactionHistory)

		// Owned Sukuk endpoint (Portfolio)
		v1.GET("/owned-sukuk/:address", handlers.GetSukukOwnedByAddress)

		// Portfolio endpoints
		v1.GET("/portfolio/:address", handlers.GetUserPortfolio)
		v1.GET("/yield-claims/:address", handlers.GetYieldClaims)
		v1.GET("/yield-distributions/:sukuk_address", handlers.GetYieldDistributions)

		// Redemption endpoints
		v1.GET("/redemptions", handlers.GetAllRedemptions)
		v1.GET("/redemptions/stats", handlers.GetRedemptionStats)
		v1.GET("/redemptions/user/:address", handlers.GetRedemptionsByUser)
		v1.GET("/redemptions/sukuk/:sukuk_address", handlers.GetRedemptionsBySukuk)
		v1.GET("/redemptions/:request_id", handlers.GetRedemptionByID)

		// Debug endpoints (optional - remove in production)
		debug := v1.Group("/debug")
		{
			debug.GET("/indexer", handlers.DebugIndexerConnection)
			debug.GET("/indexer-tables", handlers.ListIndexerTables)
			debug.GET("/indexer-tables/validate", handlers.ValidateIndexerTables)
			debug.GET("/indexer-tables/:table_name", handlers.GetTableDetails)
			debug.GET("/indexer-tables/prefix/:hash_prefix", handlers.GetHashPrefixTables)
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
