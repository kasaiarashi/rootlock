package webauthn

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/rootlock/rootlock/backend/config"
	"github.com/rootlock/rootlock/backend/internal/db"
)

var (
	ErrChallengeNotFound = errors.New("challenge not found or expired")
	ErrChallengeExpired  = errors.New("challenge has expired")
	ErrPasskeyNotFound   = errors.New("passkey not found")
	ErrUserNotFound      = errors.New("user not found")
	ErrInvalidCredential = errors.New("invalid credential")
)

// Service handles WebAuthn/passkey operations
type Service struct {
	db       *db.DB
	webauthn *webauthn.WebAuthn
}

// NewService creates a new WebAuthn service
func NewService(database *db.DB, cfg *config.Config) (*Service, error) {
	wconfig := &webauthn.Config{
		RPID:          cfg.WebAuthn.RPID,
		RPDisplayName: cfg.WebAuthn.RPName,
		RPOrigins:     cfg.WebAuthn.RPOrigins,
	}

	w, err := webauthn.New(wconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create webauthn: %w", err)
	}

	return &Service{
		db:       database,
		webauthn: w,
	}, nil
}

// BeginRegistrationRequest represents the request to begin passkey registration
type BeginRegistrationRequest struct {
	UserID     uuid.UUID `json:"user_id" binding:"required"`
	DeviceName string    `json:"device_name" binding:"required"`
}

// BeginRegistrationResponse represents the response with credential creation options
type BeginRegistrationResponse struct {
	ChallengeID uuid.UUID                    `json:"challenge_id"`
	Options     *protocol.CredentialCreation `json:"options"`
}

// FinishRegistrationRequest represents the request to complete passkey registration
type FinishRegistrationRequest struct {
	ChallengeID uuid.UUID                              `json:"challenge_id" binding:"required"`
	Credential  *protocol.ParsedCredentialCreationData `json:"credential" binding:"required"`
	DeviceName  string                                 `json:"device_name"`
}

// BeginAuthenticationRequest represents the request to begin passkey authentication
type BeginAuthenticationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// BeginAuthenticationResponse represents the response with credential request options
type BeginAuthenticationResponse struct {
	ChallengeID uuid.UUID                     `json:"challenge_id"`
	Options     *protocol.CredentialAssertion `json:"options"`
}

// FinishAuthenticationRequest represents the request to complete passkey authentication
type FinishAuthenticationRequest struct {
	ChallengeID uuid.UUID                               `json:"challenge_id" binding:"required"`
	Credential  *protocol.ParsedCredentialAssertionData `json:"credential" binding:"required"`
}

// PasskeyResponse represents a passkey in API responses
type PasskeyResponse struct {
	ID                uuid.UUID  `json:"id"`
	DeviceName        string     `json:"device_name"`
	BackupEligible    bool       `json:"backup_eligible"`
	BackupState       bool       `json:"backup_state"`
	AttestationFormat string     `json:"attestation_format,omitempty"`
	Transports        []string   `json:"transports,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	LastUsedAt        *time.Time `json:"last_used_at,omitempty"`
}

// webAuthnUser is an adapter to make db.User compatible with webauthn.User interface
type webAuthnUser struct {
	user     *db.User
	passkeys []db.Passkey
}

func (u *webAuthnUser) WebAuthnID() []byte {
	return []byte(u.user.ID.String())
}

func (u *webAuthnUser) WebAuthnName() string {
	return u.user.Email
}

func (u *webAuthnUser) WebAuthnDisplayName() string {
	return u.user.Email
}

func (u *webAuthnUser) WebAuthnCredentials() []webauthn.Credential {
	credentials := make([]webauthn.Credential, len(u.passkeys))
	for i, pk := range u.passkeys {
		var transports []protocol.AuthenticatorTransport
		for _, t := range pk.Transports {
			transports = append(transports, protocol.AuthenticatorTransport(t))
		}

		credentials[i] = webauthn.Credential{
			ID:              pk.CredentialID,
			PublicKey:       pk.CredentialPublicKey,
			AttestationType: pk.AttestationFormat,
			Transport:       transports,
			Flags: webauthn.CredentialFlags{
				BackupEligible: pk.BackupEligible,
				BackupState:    pk.BackupState,
			},
			Authenticator: webauthn.Authenticator{
				SignCount: uint32(pk.Counter),
			},
		}
	}
	return credentials
}

func (u *webAuthnUser) WebAuthnIcon() string {
	return ""
}

// BeginRegistration initiates the passkey registration process
func (s *Service) BeginRegistration(req *BeginRegistrationRequest) (*BeginRegistrationResponse, error) {
	// Find user
	var user db.User
	if err := s.db.Where("id = ?", req.UserID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Get existing passkeys for this user
	var passkeys []db.Passkey
	s.db.Where("user_id = ?", user.ID).Find(&passkeys)

	// Create WebAuthn user adapter
	wUser := &webAuthnUser{
		user:     &user,
		passkeys: passkeys,
	}

	// Create registration options
	options, sessionData, err := s.webauthn.BeginRegistration(
		wUser,
		webauthn.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			RequireResidentKey: protocol.ResidentKeyNotRequired(),
			UserVerification:   protocol.VerificationPreferred,
		}),
		webauthn.WithConveyancePreference(protocol.PreferNoAttestation),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to begin registration: %w", err)
	}

	// Store challenge in database (convert string to bytes)
	challengeBytes, err := base64.RawURLEncoding.DecodeString(sessionData.Challenge)
	if err != nil {
		return nil, fmt.Errorf("failed to decode challenge: %w", err)
	}

	challengeID := uuid.New()
	challenge := db.WebAuthnChallenge{
		ID:            challengeID,
		UserID:        &user.ID,
		Challenge:     challengeBytes,
		ChallengeType: "registration",
		ExpiresAt:     time.Now().Add(5 * time.Minute),
	}

	if err := s.db.Create(&challenge).Error; err != nil {
		return nil, fmt.Errorf("failed to store challenge: %w", err)
	}

	return &BeginRegistrationResponse{
		ChallengeID: challengeID,
		Options:     options,
	}, nil
}

// FinishRegistration completes the passkey registration process
func (s *Service) FinishRegistration(req *FinishRegistrationRequest) (*PasskeyResponse, error) {
	// Retrieve and validate challenge
	var challenge db.WebAuthnChallenge
	if err := s.db.Where("id = ? AND challenge_type = ?", req.ChallengeID, "registration").First(&challenge).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChallengeNotFound
		}
		return nil, fmt.Errorf("failed to retrieve challenge: %w", err)
	}

	if time.Now().After(challenge.ExpiresAt) {
		s.db.Delete(&challenge)
		return nil, ErrChallengeExpired
	}

	if challenge.UserID == nil {
		return nil, fmt.Errorf("challenge has no associated user")
	}

	// Get user
	var user db.User
	if err := s.db.Where("id = ?", *challenge.UserID).First(&user).Error; err != nil {
		return nil, ErrUserNotFound
	}

	// Get existing passkeys
	var passkeys []db.Passkey
	s.db.Where("user_id = ?", user.ID).Find(&passkeys)

	wUser := &webAuthnUser{
		user:     &user,
		passkeys: passkeys,
	}

	// Recreate session data from challenge (convert bytes to string)
	sessionData := webauthn.SessionData{
		Challenge:        base64.RawURLEncoding.EncodeToString(challenge.Challenge),
		UserID:           []byte(user.ID.String()),
		UserVerification: protocol.VerificationPreferred,
	}

	// Parse and verify the credential
	credential, err := s.webauthn.CreateCredential(wUser, sessionData, req.Credential)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	// Extract transports
	var transports []string
	for _, t := range credential.Transport {
		transports = append(transports, string(t))
	}

	// Extract AAGUID if available
	var aaguid *uuid.UUID
	if len(credential.Authenticator.AAGUID) == 16 {
		parsedAAGUID, err := uuid.FromBytes(credential.Authenticator.AAGUID)
		if err == nil {
			aaguid = &parsedAAGUID
		}
	}

	// Store passkey
	passkey := db.Passkey{
		ID:                  uuid.New(),
		UserID:              user.ID,
		CredentialID:        credential.ID,
		CredentialPublicKey: credential.PublicKey,
		Counter:             int64(credential.Authenticator.SignCount),
		Transports:          transports,
		BackupEligible:      credential.Flags.BackupEligible,
		BackupState:         credential.Flags.BackupState,
		AttestationFormat:   credential.AttestationType,
		AAGUID:              aaguid,
		DeviceName:          req.DeviceName,
	}

	if err := s.db.Create(&passkey).Error; err != nil {
		return nil, fmt.Errorf("failed to store passkey: %w", err)
	}

	// Delete used challenge
	s.db.Delete(&challenge)

	return &PasskeyResponse{
		ID:                passkey.ID,
		DeviceName:        passkey.DeviceName,
		BackupEligible:    passkey.BackupEligible,
		BackupState:       passkey.BackupState,
		AttestationFormat: passkey.AttestationFormat,
		Transports:        passkey.Transports,
		CreatedAt:         passkey.CreatedAt,
		LastUsedAt:        passkey.LastUsedAt,
	}, nil
}

// BeginAuthentication initiates the passkey authentication process
func (s *Service) BeginAuthentication(req *BeginAuthenticationRequest) (*BeginAuthenticationResponse, error) {
	// Find user by email
	var user db.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Get user's passkeys
	var passkeys []db.Passkey
	if err := s.db.Where("user_id = ?", user.ID).Find(&passkeys).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve passkeys: %w", err)
	}

	if len(passkeys) == 0 {
		return nil, ErrPasskeyNotFound
	}

	wUser := &webAuthnUser{
		user:     &user,
		passkeys: passkeys,
	}

	// Create authentication options
	options, sessionData, err := s.webauthn.BeginLogin(
		wUser,
		webauthn.WithUserVerification(protocol.VerificationPreferred),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to begin authentication: %w", err)
	}

	// Store challenge in database (convert string to bytes)
	challengeBytes, err := base64.RawURLEncoding.DecodeString(sessionData.Challenge)
	if err != nil {
		return nil, fmt.Errorf("failed to decode challenge: %w", err)
	}

	challengeID := uuid.New()
	challenge := db.WebAuthnChallenge{
		ID:            challengeID,
		UserID:        &user.ID,
		Challenge:     challengeBytes,
		ChallengeType: "authentication",
		ExpiresAt:     time.Now().Add(5 * time.Minute),
	}

	if err := s.db.Create(&challenge).Error; err != nil {
		return nil, fmt.Errorf("failed to store challenge: %w", err)
	}

	return &BeginAuthenticationResponse{
		ChallengeID: challengeID,
		Options:     options,
	}, nil
}

// FinishAuthentication completes the passkey authentication process
func (s *Service) FinishAuthentication(req *FinishAuthenticationRequest) (*db.User, error) {
	// Retrieve and validate challenge
	var challenge db.WebAuthnChallenge
	if err := s.db.Where("id = ? AND challenge_type = ?", req.ChallengeID, "authentication").First(&challenge).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrChallengeNotFound
		}
		return nil, fmt.Errorf("failed to retrieve challenge: %w", err)
	}

	if time.Now().After(challenge.ExpiresAt) {
		s.db.Delete(&challenge)
		return nil, ErrChallengeExpired
	}

	if challenge.UserID == nil {
		return nil, fmt.Errorf("challenge has no associated user")
	}

	// Get user
	var user db.User
	if err := s.db.Where("id = ?", *challenge.UserID).First(&user).Error; err != nil {
		return nil, ErrUserNotFound
	}

	// Get user's passkeys
	var passkeys []db.Passkey
	if err := s.db.Where("user_id = ?", user.ID).Find(&passkeys).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve passkeys: %w", err)
	}

	wUser := &webAuthnUser{
		user:     &user,
		passkeys: passkeys,
	}

	// Recreate session data from challenge (convert bytes to string)
	sessionData := webauthn.SessionData{
		Challenge:        base64.RawURLEncoding.EncodeToString(challenge.Challenge),
		UserID:           []byte(user.ID.String()),
		UserVerification: protocol.VerificationPreferred,
	}

	// Parse and verify the credential
	credential, err := s.webauthn.ValidateLogin(wUser, sessionData, req.Credential)
	if err != nil {
		return nil, fmt.Errorf("failed to validate credential: %w", err)
	}

	// Find and update the passkey with new counter
	var passkey db.Passkey
	if err := s.db.Where("user_id = ? AND credential_id = ?", user.ID, credential.ID).First(&passkey).Error; err != nil {
		return nil, ErrPasskeyNotFound
	}

	// Update counter and last used time
	now := time.Now()
	passkey.Counter = int64(credential.Authenticator.SignCount)
	passkey.LastUsedAt = &now
	if err := s.db.Save(&passkey).Error; err != nil {
		return nil, fmt.Errorf("failed to update passkey: %w", err)
	}

	// Update user's last login time
	user.LastLoginAt = &now
	s.db.Save(&user)

	// Delete used challenge
	s.db.Delete(&challenge)

	return &user, nil
}

// ListPasskeys returns all passkeys for a user
func (s *Service) ListPasskeys(userID uuid.UUID) ([]PasskeyResponse, error) {
	var passkeys []db.Passkey
	if err := s.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&passkeys).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve passkeys: %w", err)
	}

	response := make([]PasskeyResponse, len(passkeys))
	for i, pk := range passkeys {
		response[i] = PasskeyResponse{
			ID:                pk.ID,
			DeviceName:        pk.DeviceName,
			BackupEligible:    pk.BackupEligible,
			BackupState:       pk.BackupState,
			AttestationFormat: pk.AttestationFormat,
			Transports:        pk.Transports,
			CreatedAt:         pk.CreatedAt,
			LastUsedAt:        pk.LastUsedAt,
		}
	}

	return response, nil
}

// DeletePasskey deletes a passkey by ID
func (s *Service) DeletePasskey(userID, passkeyID uuid.UUID) error {
	result := s.db.Where("id = ? AND user_id = ?", passkeyID, userID).Delete(&db.Passkey{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete passkey: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrPasskeyNotFound
	}

	return nil
}

// generateChallenge generates a cryptographically secure random challenge
func generateChallenge() ([]byte, error) {
	challenge := make([]byte, 32)
	if _, err := rand.Read(challenge); err != nil {
		return nil, fmt.Errorf("failed to generate challenge: %w", err)
	}
	return challenge, nil
}
