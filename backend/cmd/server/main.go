package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rootlock/rootlock/backend/api"
	"github.com/rootlock/rootlock/backend/config"
	"github.com/rootlock/rootlock/backend/internal/db"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	database, err := db.New(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Auto-migrate in development mode only
	if cfg.IsDevelopment() {
		log.Println("Running auto-migration (development mode)")
		if err := database.AutoMigrate(); err != nil {
			log.Fatalf("Failed to auto-migrate: %v", err)
		}
	}

	// Setup router
	router := api.SetupRouter(cfg, database)

	// Create HTTP server
	server := &http.Server{
		Addr:           fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port),
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
	}

	// Start server in a goroutine
	go func() {
		log.Printf("🚀 RootLock backend starting on %s:%s", cfg.Server.Host, cfg.Server.Port)
		log.Printf("Environment: %s", cfg.Server.Environment)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
