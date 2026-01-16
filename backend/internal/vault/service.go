package vault

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/rootlock/rootlock/backend/internal/db"
)

var (
	ErrVaultNotFound   = errors.New("vault not found")
	ErrVersionConflict = errors.New("vault version conflict")
	ErrUnauthorized    = errors.New("unauthorized access")
)

// GetVaultRequest represents a request to get a vault
type GetVaultRequest struct {
	UserID uuid.UUID
}

// UpdateVaultRequest represents a request to update a vault
type UpdateVaultRequest struct {
	UserID        uuid.UUID
	EncryptedBlob []byte
	Version       int
}

// VaultResponse represents a vault in API responses
type VaultResponse struct {
	ID            uuid.UUID `json:"id"`
	EncryptedBlob string    `json:"encrypted_blob"` // Base64 encoded
	Version       int       `json:"version"`
	LastModified  string    `json:"last_modified"`
}

// Service handles vault operations
type Service struct {
	db *db.DB
}

// NewService creates a new vault service
func NewService(database *db.DB) *Service {
	return &Service{
		db: database,
	}
}

// GetVault retrieves a user's vault
func (s *Service) GetVault(req *GetVaultRequest) (*db.Vault, error) {
	var vault db.Vault
	if err := s.db.Where("user_id = ?", req.UserID).First(&vault).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVaultNotFound
		}
		return nil, fmt.Errorf("failed to get vault: %w", err)
	}

	return &vault, nil
}

// UpdateVault updates a user's vault with optimistic locking
func (s *Service) UpdateVault(req *UpdateVaultRequest) (*db.Vault, error) {
	// Start transaction for atomic update with version check
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Get current vault with lock
	var vault db.Vault
	if err := tx.Where("user_id = ?", req.UserID).First(&vault).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrVaultNotFound
		}
		return nil, fmt.Errorf("failed to get vault: %w", err)
	}

	// Check version for optimistic locking
	if vault.Version != req.Version {
		tx.Rollback()
		return nil, ErrVersionConflict
	}

	// Update vault
	vault.EncryptedBlob = req.EncryptedBlob
	vault.Version = req.Version + 1

	if err := tx.Save(&vault).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update vault: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &vault, nil
}

// CreateVault creates a new vault for a user (typically done during registration)
func (s *Service) CreateVault(userID uuid.UUID) (*db.Vault, error) {
	vault := db.Vault{
		ID:            uuid.New(),
		UserID:        userID,
		EncryptedBlob: []byte{}, // Empty vault
		Version:       1,
	}

	if err := s.db.Create(&vault).Error; err != nil {
		return nil, fmt.Errorf("failed to create vault: %w", err)
	}

	return &vault, nil
}
