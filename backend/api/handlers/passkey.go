package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/rootlock/rootlock/backend/api/middleware"
	"github.com/rootlock/rootlock/backend/internal/audit"
	"github.com/rootlock/rootlock/backend/internal/auth"
	"github.com/rootlock/rootlock/backend/internal/webauthn"
)

// PasskeyHandler handles passkey-related requests
type PasskeyHandler struct {
	webauthnService *webauthn.Service
	authService     *auth.Service
}

// NewPasskeyHandler creates a new passkey handler
func NewPasskeyHandler(webauthnService *webauthn.Service, authService *auth.Service) *PasskeyHandler {
	return &PasskeyHandler{
		webauthnService: webauthnService,
		authService:     authService,
	}
}

// BeginRegistration handles POST /api/v1/passkeys/register/begin
func (h *PasskeyHandler) BeginRegistration(c *gin.Context) {
	// Get user ID from JWT context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req webauthn.BeginRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ensure user can only register passkeys for themselves
	req.UserID = userID.(uuid.UUID)

	resp, err := h.webauthnService.BeginRegistration(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// FinishRegistration handles POST /api/v1/passkeys/register/finish
func (h *PasskeyHandler) FinishRegistration(c *gin.Context) {
	// Get user ID from JWT context for authorization
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req webauthn.FinishRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.webauthnService.FinishRegistration(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log successful passkey registration
	if auditSvc := middleware.GetAuditService(c); auditSvc != nil {
		ip, ua, uid := middleware.GetAuditInfo(c)
		auditSvc.LogSuccess(audit.EventPasskeyRegistration, uid, ip, ua, map[string]interface{}{
			"passkey_id":  resp.ID,
			"device_name": resp.DeviceName,
		})
	}

	c.JSON(http.StatusCreated, resp)
}

// BeginAuthentication handles POST /api/v1/passkeys/authenticate/begin
func (h *PasskeyHandler) BeginAuthentication(c *gin.Context) {
	var req webauthn.BeginAuthenticationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.webauthnService.BeginAuthentication(&req)
	if err != nil {
		// Log failed authentication attempt
		if auditSvc := middleware.GetAuditService(c); auditSvc != nil {
			ip, ua, _ := middleware.GetAuditInfo(c)
			auditSvc.LogFailure(audit.EventPasskeyAuthFailed, nil, ip, ua, map[string]interface{}{
				"email": req.Email,
				"error": err.Error(),
			})
		}

		if err == webauthn.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if err == webauthn.ErrPasskeyNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "no passkeys found for user"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// FinishAuthentication handles POST /api/v1/passkeys/authenticate/finish
func (h *PasskeyHandler) FinishAuthentication(c *gin.Context) {
	var req webauthn.FinishAuthenticationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.webauthnService.FinishAuthentication(&req)
	if err != nil {
		// Log failed passkey authentication
		if auditSvc := middleware.GetAuditService(c); auditSvc != nil {
			ip, ua, _ := middleware.GetAuditInfo(c)
			auditSvc.LogFailure(audit.EventPasskeyAuthFailed, nil, ip, ua, map[string]interface{}{
				"error": err.Error(),
			})
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate JWT tokens for the authenticated user
	tokens, err := h.authService.GenerateTokensForUser(user.ID, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate tokens"})
		return
	}

	// Log successful passkey authentication
	if auditSvc := middleware.GetAuditService(c); auditSvc != nil {
		ip, ua, _ := middleware.GetAuditInfo(c)
		auditSvc.LogSuccess(audit.EventPasskeyAuthentication, &user.ID, ip, ua, map[string]interface{}{
			"email": user.Email,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":         user.ID,
			"email":      user.Email,
			"created_at": user.CreatedAt,
			"last_login": user.LastLoginAt,
		},
		"tokens": tokens,
	})
}

// ListPasskeys handles GET /api/v1/passkeys
func (h *PasskeyHandler) ListPasskeys(c *gin.Context) {
	// Get user ID from JWT context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	passkeys, err := h.webauthnService.ListPasskeys(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"passkeys": passkeys})
}

// DeletePasskey handles DELETE /api/v1/passkeys/:id
func (h *PasskeyHandler) DeletePasskey(c *gin.Context) {
	// Get user ID from JWT context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	passkeyID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid passkey ID"})
		return
	}

	if err := h.webauthnService.DeletePasskey(userID.(uuid.UUID), passkeyID); err != nil {
		if err == webauthn.ErrPasskeyNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "passkey not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log passkey deletion
	if auditSvc := middleware.GetAuditService(c); auditSvc != nil {
		ip, ua, uid := middleware.GetAuditInfo(c)
		auditSvc.LogSuccess(audit.EventPasskeyDeleted, uid, ip, ua, map[string]interface{}{
			"passkey_id": passkeyID,
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "passkey deleted successfully"})
}
