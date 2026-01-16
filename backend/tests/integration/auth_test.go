package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rootlock/rootlock/backend/api"
	"github.com/rootlock/rootlock/backend/config"
	"github.com/rootlock/rootlock/backend/internal/db"
)

// setupTestRouter creates a test router with in-memory database
func setupTestRouter(t *testing.T) (*http.Handler, *db.DB, func()) {
	// Load test config
	cfg := &config.Config{
		Server: config.ServerConfig{
			Port:        "8080",
			Host:        "0.0.0.0",
			Environment: "test",
			CORSOrigins: []string{"*"},
		},
		Database: config.DatabaseConfig{
			URL:          "postgresql://rootlock:rootlock_dev_password@localhost:5432/rootlock_test?sslmode=disable",
			MaxOpenConns: 5,
			MaxIdleConns: 2,
		},
		Redis: config.RedisConfig{
			URL: "redis://localhost:6379/1",
		},
		JWT: config.JWTConfig{
			Secret:               "test-secret-key",
			AccessTokenDuration:  900,
			RefreshTokenDuration: 2592000,
		},
		WebAuthn: config.WebAuthnConfig{
			RPID:      "localhost",
			RPName:    "RootLock Test",
			RPOrigins: []string{"http://localhost:8080"},
		},
		Security: config.SecurityConfig{
			BcryptCost:         10,
			RateLimitEnabled:   false, // Disable for tests
			RateLimitPerMinute: 60,
			Argon2Memory:       65536,
			Argon2Time:         3,
			Argon2Parallelism:  4,
		},
	}

	// Connect to test database
	database, err := db.Connect(cfg.Database.URL)
	require.NoError(t, err)

	// Run migrations
	err = database.AutoMigrate(
		&db.User{},
		&db.Vault{},
		&db.Device{},
		&db.Passkey{},
		&db.AuditLog{},
		&db.RateLimit{},
		&db.WebAuthnChallenge{},
		&db.Session{},
	)
	require.NoError(t, err)

	// Setup router
	router := api.SetupRouter(cfg, database)

	cleanup := func() {
		// Clean up test data
		database.Exec("DROP SCHEMA public CASCADE")
		database.Exec("CREATE SCHEMA public")
		db.Close(database)
	}

	var handler http.Handler = router
	return &handler, database, cleanup
}

func TestUserRegistration(t *testing.T) {
	handler, _, cleanup := setupTestRouter(t)
	defer cleanup()

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name: "successful registration",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"auk_hash": "test-auk-hash-value",
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body, "user")
				user := body["user"].(map[string]interface{})
				assert.Equal(t, "test@example.com", user["email"])
				assert.Contains(t, user, "id")
			},
		},
		{
			name: "duplicate email",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"auk_hash": "test-auk-hash-value-2",
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body, "error")
			},
		},
		{
			name: "missing email",
			requestBody: map[string]interface{}{
				"auk_hash": "test-auk-hash-value",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body, "error")
			},
		},
		{
			name: "invalid email",
			requestBody: map[string]interface{}{
				"email":    "invalid-email",
				"auk_hash": "test-auk-hash-value",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body, "error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			(*handler).ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.checkResponse != nil {
				tt.checkResponse(t, response)
			}
		})
	}
}

func TestUserLogin(t *testing.T) {
	handler, _, cleanup := setupTestRouter(t)
	defer cleanup()

	// First, register a user
	registerBody := map[string]interface{}{
		"email":    "login@example.com",
		"auk_hash": "test-auk-for-login",
	}
	jsonBody, _ := json.Marshal(registerBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	(*handler).ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name: "successful login",
			requestBody: map[string]interface{}{
				"email": "login@example.com",
				"auk":   "test-auk-for-login",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body, "user")
				assert.Contains(t, body, "tokens")
				tokens := body["tokens"].(map[string]interface{})
				assert.Contains(t, tokens, "access_token")
				assert.Contains(t, tokens, "refresh_token")
			},
		},
		{
			name: "invalid credentials",
			requestBody: map[string]interface{}{
				"email": "login@example.com",
				"auk":   "wrong-auk",
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body, "error")
			},
		},
		{
			name: "non-existent user",
			requestBody: map[string]interface{}{
				"email": "nonexistent@example.com",
				"auk":   "some-auk",
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body, "error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(jsonBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			(*handler).ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.checkResponse != nil {
				tt.checkResponse(t, response)
			}
		})
	}
}

func TestHealthEndpoints(t *testing.T) {
	handler, _, cleanup := setupTestRouter(t)
	defer cleanup()

	tests := []struct {
		name           string
		endpoint       string
		expectedStatus int
	}{
		{
			name:           "health check",
			endpoint:       "/health",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "readiness check",
			endpoint:       "/ready",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.endpoint, nil)
			w := httptest.NewRecorder()

			(*handler).ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.Equal(t, "ok", response["status"])
		})
	}
}
