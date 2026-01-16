package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/rootlock/rootlock/backend/config"
	"github.com/rootlock/rootlock/backend/internal/db"
)

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
)

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email   string `json:"email" binding:"required,email"`
	AUKHash string `json:"auk_hash" binding:"required"` // Pre-hashed on client
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email string `json:"email" binding:"required,email"`
	AUK   string `json:"auk" binding:"required"` // Account Unlock Key (not password!)
}

// LoginResponse represents a login response
type LoginResponse struct {
	User   *UserResponse `json:"user"`
	Tokens *TokenPair    `json:"tokens"`
}

// UserResponse represents a user in API responses
type UserResponse struct {
	ID        uuid.UUID  `json:"id"`
	Email     string     `json:"email"`
	CreatedAt time.Time  `json:"created_at"`
	LastLogin *time.Time `json:"last_login,omitempty"`
}

// Service handles authentication operations
type Service struct {
	db         *db.DB
	jwtService *JWTService
	bcryptCost int
}

// NewService creates a new authentication service
func NewService(database *db.DB, cfg *config.Config) *Service {
	jwtService := NewJWTService(
		cfg.JWT.Secret,
		cfg.JWT.AccessTokenDuration,
		cfg.JWT.RefreshTokenDuration,
	)

	return &Service{
		db:         database,
		jwtService: jwtService,
		bcryptCost: cfg.Security.BcryptCost,
	}
}

// Register registers a new user
// Note: AUK is already derived client-side via HKDF(MEK, "account-unlock-key")
// We just need to hash it with bcrypt for storage
func (s *Service) Register(req *RegisterRequest) (*UserResponse, error) {
	// Check if user already exists
	var existingUser db.User
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash the AUK with bcrypt for secure storage
	// Note: The client sends a pre-derived AUK, we just add bcrypt layer
	aukHashBytes, err := bcrypt.GenerateFromPassword([]byte(req.AUKHash), s.bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash AUK: %w", err)
	}

	// Create user
	user := db.User{
		ID:      uuid.New(),
		Email:   req.Email,
		AUKHash: string(aukHashBytes),
	}

	if err := s.db.Create(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create empty vault for user
	vault := db.Vault{
		ID:            uuid.New(),
		UserID:        user.ID,
		EncryptedBlob: []byte{}, // Empty vault initially
		Version:       1,
	}

	if err := s.db.Create(&vault).Error; err != nil {
		// Rollback user creation if vault creation fails
		s.db.Delete(&user)
		return nil, fmt.Errorf("failed to create vault: %w", err)
	}

	return &UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}, nil
}

// Login authenticates a user and returns tokens
func (s *Service) Login(req *LoginRequest) (*LoginResponse, error) {
	// Find user by email
	var user db.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Verify AUK using constant-time comparison (bcrypt does this)
	if err := bcrypt.CompareHashAndPassword([]byte(user.AUKHash), []byte(req.AUK)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Update last login time
	now := time.Now()
	user.LastLoginAt = &now
	s.db.Save(&user)

	// Generate JWT tokens
	tokens, err := s.jwtService.GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return &LoginResponse{
		User: &UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			LastLogin: user.LastLoginAt,
		},
		Tokens: tokens,
	}, nil
}

// RefreshToken generates a new access token from a refresh token
func (s *Service) RefreshToken(refreshToken string) (string, error) {
	return s.jwtService.RefreshAccessToken(refreshToken)
}

// ValidateToken validates a JWT token and returns the claims
func (s *Service) ValidateToken(token string) (*Claims, error) {
	return s.jwtService.ValidateToken(token)
}

// GetUserByID retrieves a user by ID
func (s *Service) GetUserByID(userID uuid.UUID) (*UserResponse, error) {
	var user db.User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
		LastLogin: user.LastLoginAt,
	}, nil
}
