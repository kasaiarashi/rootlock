package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/rootlock/rootlock/backend/internal/db"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	db *db.DB
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(database *db.DB) *HealthHandler {
	return &HealthHandler{
		db: database,
	}
}

// Health returns the health status of the service
// GET /health
func (h *HealthHandler) Health(c *gin.Context) {
	// Check database connection
	if err := h.db.Health(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":   "unhealthy",
			"database": "disconnected",
			"error":    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":   "healthy",
		"database": "connected",
	})
}

// Ready returns readiness status
// GET /ready
func (h *HealthHandler) Ready(c *gin.Context) {
	if err := h.db.Health(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"ready": false,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ready": true,
	})
}
