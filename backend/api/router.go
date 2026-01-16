package api

import (
	"github.com/gin-gonic/gin"

	"github.com/rootlock/rootlock/backend/api/handlers"
	"github.com/rootlock/rootlock/backend/api/middleware"
	"github.com/rootlock/rootlock/backend/config"
	"github.com/rootlock/rootlock/backend/internal/auth"
	"github.com/rootlock/rootlock/backend/internal/db"
	"github.com/rootlock/rootlock/backend/internal/vault"
)

// SetupRouter creates and configures the Gin router
func SetupRouter(cfg *config.Config, database *db.DB) *gin.Engine {
	// Set Gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.CORSMiddleware(cfg.Server.CORSOrigins))

	// Initialize services
	authService := auth.NewService(database, cfg)
	vaultService := vault.NewService(database)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(database)
	authHandler := handlers.NewAuthHandler(authService)
	vaultHandler := handlers.NewVaultHandler(vaultService)

	// Health check endpoints (no auth required)
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public auth routes (no auth required)
		authRoutes := v1.Group("/auth")
		{
			authRoutes.POST("/register", authHandler.Register)
			authRoutes.POST("/login", authHandler.Login)
			authRoutes.POST("/refresh", authHandler.RefreshToken)
		}

		// Protected routes (require authentication)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(authService))
		{
			// Auth routes
			protected.GET("/auth/me", authHandler.GetMe)

			// Vault routes
			protected.GET("/vault", vaultHandler.GetVault)
			protected.PUT("/vault", vaultHandler.UpdateVault)
		}
	}

	return router
}
