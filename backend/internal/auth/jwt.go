package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims represents the JWT claims
type Claims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

// TokenPair represents an access and refresh token pair
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // seconds
}

// JWTService handles JWT token operations
type JWTService struct {
	secret               string
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

// NewJWTService creates a new JWT service
func NewJWTService(secret string, accessDuration, refreshDuration time.Duration) *JWTService {
	return &JWTService{
		secret:               secret,
		accessTokenDuration:  accessDuration,
		refreshTokenDuration: refreshDuration,
	}
}

// GenerateTokenPair generates a new access and refresh token pair
func (s *JWTService) GenerateTokenPair(userID uuid.UUID, email string) (*TokenPair, error) {
	// Generate access token
	accessToken, err := s.generateToken(userID, email, s.accessTokenDuration)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := s.generateToken(userID, email, s.refreshTokenDuration)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.accessTokenDuration.Seconds()),
	}, nil
}

// generateToken generates a JWT token
func (s *JWTService) generateToken(userID uuid.UUID, email string, duration time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "rootlock",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secret))
}

// ValidateToken validates a JWT token and returns the claims
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshAccessToken generates a new access token from a valid refresh token
func (s *JWTService) RefreshAccessToken(refreshToken string) (string, error) {
	// Validate refresh token
	claims, err := s.ValidateToken(refreshToken)
	if err != nil {
		return "", err
	}

	// Generate new access token
	return s.generateToken(claims.UserID, claims.Email, s.accessTokenDuration)
}
