package api

import (
	"github.com/gin-gonic/gin"

	"github.com/rootlock/rootlock/backend/api/handlers"
	"github.com/rootlock/rootlock/backend/api/middleware"
	"github.com/rootlock/rootlock/backend/config"
	"github.com/rootlock/rootlock/backend/internal/audit"
	"github.com/rootlock/rootlock/backend/internal/auth"
	"github.com/rootlock/rootlock/backend/internal/db"
	"github.com/rootlock/rootlock/backend/internal/vault"
	"github.com/rootlock/rootlock/backend/internal/webauthn"
)

// SetupRouter creates and configures the Gin router
func SetupRouter(cfg *config.Config, database *db.DB) *gin.Engine {
	// Set Gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Initialize services
	authService := auth.NewService(database, cfg)
	vaultService := vault.NewService(database)
	auditService := audit.NewService(database)
	webauthnService, err2 := webauthn.NewService(database, cfg)
	if err2 != nil {
		panic("failed to initialize webauthn service: " + err2.Error())
	}

	// Initialize rate limiter
	rateLimiter, err := middleware.NewRateLimiter(cfg)
	if err != nil {
		panic("failed to initialize rate limiter: " + err.Error())
	}

	// Global middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.SecurityHeaders())
	router.Use(middleware.CORSMiddleware(cfg.Server.CORSOrigins))
	router.Use(middleware.AuditMiddleware(auditService)) // Add audit context
	router.Use(rateLimiter.Middleware())                 // Apply rate limiting globally

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(database)
	authHandler := handlers.NewAuthHandler(authService)
	vaultHandler := handlers.NewVaultHandler(vaultService)
	passkeyHandler := handlers.NewPasskeyHandler(webauthnService, authService)

	// Health check endpoints (no auth required)
	router.GET("/health", healthHandler.Health)
	router.GET("/ready", healthHandler.Ready)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public auth routes (no auth required, with strict rate limiting)
		authRoutes := v1.Group("/auth")
		{
			// Strict rate limiting for registration (5 per minute)
			authRoutes.POST("/register", middleware.StrictRateLimitMiddleware(rateLimiter, 5), authHandler.Register)
			// Strict rate limiting for login (10 per minute)
			authRoutes.POST("/login", middleware.StrictRateLimitMiddleware(rateLimiter, 10), authHandler.Login)
			authRoutes.POST("/refresh", authHandler.RefreshToken)
		}

		// Public passkey authentication routes (no auth required, with strict rate limiting)
		passkeyAuthRoutes := v1.Group("/passkeys")
		{
			// Strict rate limiting for passkey authentication (10 per minute)
			passkeyAuthRoutes.POST("/authenticate/begin", middleware.StrictRateLimitMiddleware(rateLimiter, 10), passkeyHandler.BeginAuthentication)
			passkeyAuthRoutes.POST("/authenticate/finish", middleware.StrictRateLimitMiddleware(rateLimiter, 10), passkeyHandler.FinishAuthentication)
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

			// Passkey management routes (protected)
			passkeyRoutes := protected.Group("/passkeys")
			{
				passkeyRoutes.POST("/register/begin", passkeyHandler.BeginRegistration)
				passkeyRoutes.POST("/register/finish", passkeyHandler.FinishRegistration)
				passkeyRoutes.GET("", passkeyHandler.ListPasskeys)
				passkeyRoutes.DELETE("/:id", passkeyHandler.DeletePasskey)
			}
		}
	}

	return router
}
