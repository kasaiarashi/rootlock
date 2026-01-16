package handlers

import (
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/rootlock/rootlock/backend/api/middleware"
	"github.com/rootlock/rootlock/backend/internal/audit"
	"github.com/rootlock/rootlock/backend/internal/vault"
)

// VaultHandler handles vault endpoints
type VaultHandler struct {
	vaultService *vault.Service
}

// NewVaultHandler creates a new vault handler
func NewVaultHandler(vaultService *vault.Service) *VaultHandler {
	return &VaultHandler{
		vaultService: vaultService,
	}
}

// GetVault retrieves the user's encrypted vault
// GET /api/v1/vault
func (h *VaultHandler) GetVault(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userVault, err := h.vaultService.GetVault(&vault.GetVaultRequest{
		UserID: userID,
	})
	if err != nil {
		if err == vault.ErrVaultNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "vault not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get vault"})
		}
		return
	}

	// Log vault access
	if auditSvc := middleware.GetAuditService(c); auditSvc != nil {
		ip, ua, uid := middleware.GetAuditInfo(c)
		auditSvc.LogSuccess(audit.EventVaultAccess, uid, ip, ua, map[string]interface{}{
			"vault_id": userVault.ID,
			"version":  userVault.Version,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             userVault.ID,
		"encrypted_blob": base64.StdEncoding.EncodeToString(userVault.EncryptedBlob),
		"version":        userVault.Version,
		"last_modified":  userVault.LastModified,
	})
}

// UpdateVault updates the user's encrypted vault
// PUT /api/v1/vault
func (h *VaultHandler) UpdateVault(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		EncryptedBlob string `json:"encrypted_blob" binding:"required"` // Base64 encoded
		Version       int    `json:"version" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Decode base64 encrypted blob
	encryptedBlob, err := base64.StdEncoding.DecodeString(req.EncryptedBlob)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid encrypted_blob format"})
		return
	}

	updatedVault, err := h.vaultService.UpdateVault(&vault.UpdateVaultRequest{
		UserID:        userID,
		EncryptedBlob: encryptedBlob,
		Version:       req.Version,
	})
	if err != nil {
		// Log failed vault update
		if auditSvc := middleware.GetAuditService(c); auditSvc != nil {
			ip, ua, uid := middleware.GetAuditInfo(c)
			auditSvc.LogFailure(audit.EventVaultUpdate, uid, ip, ua, map[string]interface{}{
				"error":   err.Error(),
				"version": req.Version,
			})
		}

		if err == vault.ErrVaultNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "vault not found"})
		} else if err == vault.ErrVersionConflict {
			c.JSON(http.StatusConflict, gin.H{"error": "version conflict"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update vault"})
		}
		return
	}

	// Log successful vault update
	if auditSvc := middleware.GetAuditService(c); auditSvc != nil {
		ip, ua, uid := middleware.GetAuditInfo(c)
		auditSvc.LogSuccess(audit.EventVaultUpdate, uid, ip, ua, map[string]interface{}{
			"vault_id":    updatedVault.ID,
			"new_version": updatedVault.Version,
			"size_bytes":  len(encryptedBlob),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"id":            updatedVault.ID,
		"version":       updatedVault.Version,
		"last_modified": updatedVault.LastModified,
		"message":       "vault updated successfully",
	})
}
