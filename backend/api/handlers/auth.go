package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/rootlock/rootlock/backend/api/middleware"
	"github.com/rootlock/rootlock/backend/internal/audit"
	"github.com/rootlock/rootlock/backend/internal/auth"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *auth.Service
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *auth.Service) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req auth.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.Register(&req)
	if err != nil {
		// Log failed registration
		if auditSvc := middleware.GetAuditService(c); auditSvc != nil {
			ip, ua, _ := middleware.GetAuditInfo(c)
			auditSvc.LogFailure(audit.EventUserRegistration, nil, ip, ua, map[string]interface{}{
				"email": req.Email,
				"error": err.Error(),
			})
		}

		switch err {
		case auth.ErrUserAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
		}
		return
	}

	// Log successful registration
	if auditSvc := middleware.GetAuditService(c); auditSvc != nil {
		ip, ua, _ := middleware.GetAuditInfo(c)
		auditSvc.LogSuccess(audit.EventUserRegistration, &user.ID, ip, ua, map[string]interface{}{
			"email": user.Email,
		})
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "user registered successfully",
		"user":    user,
	})
}

// Login handles user login
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req auth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.authService.Login(&req)
	if err != nil {
		// Log failed login attempt
		if auditSvc := middleware.GetAuditService(c); auditSvc != nil {
			ip, ua, _ := middleware.GetAuditInfo(c)
			auditSvc.LogFailure(audit.EventUserLoginFailed, nil, ip, ua, map[string]interface{}{
				"email": req.Email,
				"error": err.Error(),
			})
		}

		switch err {
		case auth.ErrUserNotFound, auth.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
		}
		return
	}

	// Log successful login
	if auditSvc := middleware.GetAuditService(c); auditSvc != nil {
		ip, ua, _ := middleware.GetAuditInfo(c)
		auditSvc.LogSuccess(audit.EventUserLogin, &response.User.ID, ip, ua, map[string]interface{}{
			"email": response.User.Email,
		})
	}

	c.JSON(http.StatusOK, response)
}

// RefreshToken handles token refresh
// POST /api/v1/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		// Log failed token refresh
		if auditSvc := middleware.GetAuditService(c); auditSvc != nil {
			ip, ua, userID := middleware.GetAuditInfo(c)
			auditSvc.LogFailure(audit.EventTokenRefresh, userID, ip, ua, map[string]interface{}{
				"error": err.Error(),
			})
		}

		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired refresh token"})
		return
	}

	// Log successful token refresh
	if auditSvc := middleware.GetAuditService(c); auditSvc != nil {
		ip, ua, userID := middleware.GetAuditInfo(c)
		auditSvc.LogSuccess(audit.EventTokenRefresh, userID, ip, ua, nil)
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
	})
}

// GetMe returns the current user's information
// GET /api/v1/auth/me
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, _ := c.Get("user_id")

	user, err := h.authService.GetUserByID(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	c.JSON(http.StatusOK, user)
}
