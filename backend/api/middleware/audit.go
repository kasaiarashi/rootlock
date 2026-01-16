package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/rootlock/rootlock/backend/internal/audit"
)

// AuditMiddleware adds audit logging context to requests
func AuditMiddleware(auditService *audit.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Store audit service in context for handlers to use
		c.Set("audit_service", auditService)

		// Extract common audit information
		c.Set("audit_ip", c.ClientIP())
		c.Set("audit_user_agent", c.Request.UserAgent())

		c.Next()
	}
}

// GetAuditInfo extracts audit information from the request context
func GetAuditInfo(c *gin.Context) (ipAddress, userAgent string, userID *uuid.UUID) {
	ipAddress = c.ClientIP()
	userAgent = c.Request.UserAgent()

	if id, exists := c.Get("user_id"); exists {
		if uid, ok := id.(uuid.UUID); ok {
			userID = &uid
		}
	}

	return
}

// GetAuditService retrieves the audit service from the context
func GetAuditService(c *gin.Context) *audit.Service {
	if service, exists := c.Get("audit_service"); exists {
		if auditService, ok := service.(*audit.Service); ok {
			return auditService
		}
	}
	return nil
}
