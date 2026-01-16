package audit

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/rootlock/rootlock/backend/internal/db"
)

// Event types
const (
	EventUserRegistration      = "user.registration"
	EventUserLogin             = "user.login"
	EventUserLoginFailed       = "user.login.failed"
	EventUserLogout            = "user.logout"
	EventPasswordChange        = "user.password.change"
	EventVaultAccess           = "vault.access"
	EventVaultUpdate           = "vault.update"
	EventPasskeyRegistration   = "passkey.registration"
	EventPasskeyAuthentication = "passkey.authentication"
	EventPasskeyAuthFailed     = "passkey.authentication.failed"
	EventPasskeyDeleted        = "passkey.deleted"
	EventTokenRefresh          = "token.refresh"
	EventRateLimitExceeded     = "ratelimit.exceeded"
	EventUnauthorizedAccess    = "security.unauthorized"
	EventSuspiciousActivity    = "security.suspicious"
)

// Event statuses
const (
	StatusSuccess = "success"
	StatusFailure = "failure"
	StatusWarning = "warning"
)

// Service handles audit logging operations
type Service struct {
	db *db.DB
}

// NewService creates a new audit service
func NewService(database *db.DB) *Service {
	return &Service{
		db: database,
	}
}

// LogEvent logs an audit event
func (s *Service) LogEvent(event *AuditEvent) error {
	auditLog := db.AuditLog{
		ID:          uuid.New(),
		UserID:      event.UserID,
		EventType:   event.EventType,
		EventStatus: event.EventStatus,
		IPAddress:   event.IPAddress,
		UserAgent:   event.UserAgent,
		DeviceID:    event.DeviceID,
		Metadata:    event.Metadata,
		CreatedAt:   time.Now(),
	}

	if err := s.db.Create(&auditLog).Error; err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// LogSuccess logs a successful event
func (s *Service) LogSuccess(eventType string, userID *uuid.UUID, ipAddress, userAgent string, metadata map[string]interface{}) error {
	return s.LogEvent(&AuditEvent{
		UserID:      userID,
		EventType:   eventType,
		EventStatus: StatusSuccess,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Metadata:    metadata,
	})
}

// LogFailure logs a failed event
func (s *Service) LogFailure(eventType string, userID *uuid.UUID, ipAddress, userAgent string, metadata map[string]interface{}) error {
	return s.LogEvent(&AuditEvent{
		UserID:      userID,
		EventType:   eventType,
		EventStatus: StatusFailure,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Metadata:    metadata,
	})
}

// LogWarning logs a warning event
func (s *Service) LogWarning(eventType string, userID *uuid.UUID, ipAddress, userAgent string, metadata map[string]interface{}) error {
	return s.LogEvent(&AuditEvent{
		UserID:      userID,
		EventType:   eventType,
		EventStatus: StatusWarning,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Metadata:    metadata,
	})
}

// GetUserAuditLogs retrieves audit logs for a specific user
func (s *Service) GetUserAuditLogs(userID uuid.UUID, limit int, offset int) ([]db.AuditLog, error) {
	var logs []db.AuditLog
	query := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset)

	if err := query.Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve audit logs: %w", err)
	}

	return logs, nil
}

// GetRecentAuditLogs retrieves recent audit logs (admin function)
func (s *Service) GetRecentAuditLogs(limit int, offset int) ([]db.AuditLog, error) {
	var logs []db.AuditLog
	query := s.db.Order("created_at DESC").
		Limit(limit).
		Offset(offset)

	if err := query.Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve audit logs: %w", err)
	}

	return logs, nil
}

// GetAuditLogsByEventType retrieves audit logs by event type
func (s *Service) GetAuditLogsByEventType(eventType string, limit int, offset int) ([]db.AuditLog, error) {
	var logs []db.AuditLog
	query := s.db.Where("event_type = ?", eventType).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset)

	if err := query.Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve audit logs: %w", err)
	}

	return logs, nil
}

// GetFailedLoginAttempts retrieves failed login attempts for an IP or user
func (s *Service) GetFailedLoginAttempts(identifier string, since time.Time) (int64, error) {
	var count int64
	query := s.db.Model(&db.AuditLog{}).
		Where("event_type IN (?, ?) AND event_status = ? AND created_at >= ?",
			EventUserLoginFailed, EventPasskeyAuthFailed, StatusFailure, since)

	// Check if identifier is a UUID (user ID) or IP address
	if _, err := uuid.Parse(identifier); err == nil {
		userID, _ := uuid.Parse(identifier)
		query = query.Where("user_id = ?", userID)
	} else {
		query = query.Where("ip_address = ?", identifier)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count failed login attempts: %w", err)
	}

	return count, nil
}

// CleanupOldLogs removes audit logs older than the specified duration
func (s *Service) CleanupOldLogs(olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	result := s.db.Where("created_at < ?", cutoff).Delete(&db.AuditLog{})

	if result.Error != nil {
		return 0, fmt.Errorf("failed to cleanup old logs: %w", result.Error)
	}

	return result.RowsAffected, nil
}

// AuditEvent represents an event to be logged
type AuditEvent struct {
	UserID      *uuid.UUID
	EventType   string
	EventStatus string
	IPAddress   string
	UserAgent   string
	DeviceID    *uuid.UUID
	Metadata    map[string]interface{}
}
