package db

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user account in the system
type User struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Email       string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	AUKHash     string     `gorm:"type:varchar(255);not null" json:"-"` // Never expose in JSON
	CreatedAt   time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`

	// Relationships
	Vault    *Vault    `gorm:"foreignKey:UserID" json:"-"`
	Devices  []Device  `gorm:"foreignKey:UserID" json:"-"`
	Passkeys []Passkey `gorm:"foreignKey:UserID" json:"-"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}

// Vault represents an encrypted vault for a user
type Vault struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	EncryptedBlob []byte    `gorm:"type:bytea;not null" json:"-"` // Never expose in JSON
	Version       int       `gorm:"not null;default:1" json:"version"`
	LastModified  time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"last_modified"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// TableName specifies the table name for Vault
func (Vault) TableName() string {
	return "vaults"
}

// Device represents a user's device
type Device struct {
	ID                uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID            uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	DeviceName        string    `gorm:"type:varchar(255);not null" json:"device_name"`
	DeviceType        string    `gorm:"type:varchar(50);not null" json:"device_type"` // ios, android, desktop, browser, cli
	DeviceFingerprint string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"-"`
	Trusted           bool      `gorm:"not null;default:false" json:"trusted"`
	CreatedAt         time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	LastSeen          time.Time `gorm:"not null;default:CURRENT_TIMESTAMP" json:"last_seen"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// TableName specifies the table name for Device
func (Device) TableName() string {
	return "devices"
}

// Passkey represents a WebAuthn passkey credential
type Passkey struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID              uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	CredentialID        []byte     `gorm:"type:bytea;uniqueIndex;not null" json:"-"`
	CredentialPublicKey []byte     `gorm:"type:bytea;not null" json:"-"`
	Counter             int64      `gorm:"not null;default:0" json:"counter"`
	Transports          []string   `gorm:"type:text[]" json:"transports,omitempty"`
	BackupEligible      bool       `gorm:"not null;default:false" json:"backup_eligible"`
	BackupState         bool       `gorm:"not null;default:false" json:"backup_state"`
	AttestationFormat   string     `gorm:"type:varchar(50)" json:"attestation_format,omitempty"`
	AttestationObject   []byte     `gorm:"type:bytea" json:"-"`
	AAGUID              *uuid.UUID `gorm:"type:uuid" json:"aaguid,omitempty"`
	DeviceName          string     `gorm:"type:varchar(255)" json:"device_name,omitempty"`
	CreatedAt           time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	LastUsedAt          *time.Time `json:"last_used_at,omitempty"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// TableName specifies the table name for Passkey
func (Passkey) TableName() string {
	return "passkeys"
}

// AuditLog represents a security event log entry
type AuditLog struct {
	ID          uuid.UUID              `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID      *uuid.UUID             `gorm:"type:uuid" json:"user_id,omitempty"`
	EventType   string                 `gorm:"type:varchar(100);not null" json:"event_type"`
	EventStatus string                 `gorm:"type:varchar(20);not null" json:"event_status"` // success, failure, warning
	IPAddress   string                 `gorm:"type:inet" json:"ip_address,omitempty"`
	UserAgent   string                 `gorm:"type:text" json:"user_agent,omitempty"`
	DeviceID    *uuid.UUID             `gorm:"type:uuid" json:"device_id,omitempty"`
	Metadata    map[string]interface{} `gorm:"type:jsonb" json:"metadata,omitempty"`
	CreatedAt   time.Time              `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

// TableName specifies the table name for AuditLog
func (AuditLog) TableName() string {
	return "audit_logs"
}

// RateLimit represents a rate limiting entry
type RateLimit struct {
	ID           int        `gorm:"primaryKey;autoIncrement" json:"id"`
	Identifier   string     `gorm:"type:varchar(255);not null" json:"identifier"`
	Endpoint     string     `gorm:"type:varchar(255);not null" json:"endpoint"`
	Attempts     int        `gorm:"not null;default:1" json:"attempts"`
	WindowStart  time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"window_start"`
	BlockedUntil *time.Time `json:"blocked_until,omitempty"`
}

// TableName specifies the table name for RateLimit
func (RateLimit) TableName() string {
	return "rate_limits"
}

// WebAuthnChallenge represents a temporary WebAuthn challenge
type WebAuthnChallenge struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID        *uuid.UUID `gorm:"type:uuid" json:"user_id,omitempty"`
	Challenge     []byte     `gorm:"type:bytea;not null" json:"-"`
	ChallengeType string     `gorm:"type:varchar(20);not null" json:"challenge_type"` // registration, authentication
	ExpiresAt     time.Time  `gorm:"not null" json:"expires_at"`
	CreatedAt     time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
}

// TableName specifies the table name for WebAuthnChallenge
func (WebAuthnChallenge) TableName() string {
	return "webauthn_challenges"
}

// Session represents a user session
type Session struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID           uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	DeviceID         *uuid.UUID `gorm:"type:uuid" json:"device_id,omitempty"`
	RefreshTokenHash string     `gorm:"type:varchar(255);not null;index" json:"-"`
	ExpiresAt        time.Time  `gorm:"not null" json:"expires_at"`
	CreatedAt        time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"created_at"`
	LastUsedAt       time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP" json:"last_used_at"`
	Revoked          bool       `gorm:"not null;default:false" json:"revoked"`

	// Relationships
	User   User    `gorm:"foreignKey:UserID" json:"-"`
	Device *Device `gorm:"foreignKey:DeviceID" json:"-"`
}

// TableName specifies the table name for Session
func (Session) TableName() string {
	return "sessions"
}
