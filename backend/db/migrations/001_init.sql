-- RootLock Database Schema
-- PostgreSQL 16+
--
-- This schema implements a zero-knowledge architecture where the server
-- cannot decrypt user vault data.

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
-- Stores minimal user information required for authentication
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    auk_hash VARCHAR(255) NOT NULL,  -- Bcrypt hash of Account Unlock Key
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP,
    
    -- Indexes
    CONSTRAINT email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}$')
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_created_at ON users(created_at);

-- Vaults table
-- Stores encrypted vault blobs (zero-knowledge)
CREATE TABLE vaults (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    encrypted_blob BYTEA NOT NULL,     -- Entire vault encrypted with MEK
    version INTEGER NOT NULL DEFAULT 1, -- For conflict resolution
    last_modified TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT one_vault_per_user UNIQUE(user_id)
);

CREATE INDEX idx_vaults_user_id ON vaults(user_id);
CREATE INDEX idx_vaults_last_modified ON vaults(last_modified);

-- Devices table
-- Tracks user devices for multi-device sync
CREATE TABLE devices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_name VARCHAR(255) NOT NULL,
    device_type VARCHAR(50) NOT NULL, -- 'ios', 'desktop', 'browser', 'cli'
    device_fingerprint VARCHAR(255) UNIQUE NOT NULL, -- Unique identifier for device
    trusted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT valid_device_type CHECK (device_type IN ('ios', 'android', 'desktop', 'browser', 'cli'))
);

CREATE INDEX idx_devices_user_id ON devices(user_id);
CREATE INDEX idx_devices_fingerprint ON devices(device_fingerprint);
CREATE INDEX idx_devices_last_seen ON devices(last_seen);

-- Passkeys table (WebAuthn credentials)
-- Stores public keys for passkey authentication
CREATE TABLE passkeys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    credential_id BYTEA UNIQUE NOT NULL,          -- WebAuthn credential ID
    credential_public_key BYTEA NOT NULL,         -- Public key (never private key!)
    counter BIGINT NOT NULL DEFAULT 0,            -- Signature counter (prevent cloning)
    transports TEXT[],                            -- ['usb', 'nfc', 'ble', 'internal']
    backup_eligible BOOLEAN NOT NULL DEFAULT FALSE,
    backup_state BOOLEAN NOT NULL DEFAULT FALSE,
    attestation_format VARCHAR(50),               -- 'packed', 'tpm', 'android-key', etc.
    attestation_object BYTEA,                     -- Optional: full attestation
    aaguid UUID,                                  -- Authenticator AAGUID
    device_name VARCHAR(255),                     -- User-friendly name
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP,
    
    -- Constraints
    CONSTRAINT credential_id_unique UNIQUE(user_id, credential_id)
);

CREATE INDEX idx_passkeys_user_id ON passkeys(user_id);
CREATE INDEX idx_passkeys_credential_id ON passkeys(credential_id);
CREATE INDEX idx_passkeys_last_used ON passkeys(last_used_at);

-- Audit logs table
-- Security event logging for monitoring and compliance
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    event_type VARCHAR(100) NOT NULL,   -- 'login', 'logout', 'vault_access', etc.
    event_status VARCHAR(20) NOT NULL,  -- 'success', 'failure'
    ip_address INET,                    -- Client IP address
    user_agent TEXT,                    -- Browser/client user agent
    device_id UUID REFERENCES devices(id) ON DELETE SET NULL,
    metadata JSONB,                     -- Additional event-specific data
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT valid_event_status CHECK (event_status IN ('success', 'failure', 'warning'))
);

CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_event_type ON audit_logs(event_type);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX idx_audit_logs_metadata ON audit_logs USING GIN(metadata);

-- Rate limiting table
-- Prevents brute-force attacks
CREATE TABLE rate_limits (
    id SERIAL PRIMARY KEY,
    identifier VARCHAR(255) NOT NULL,   -- IP address or user_id
    endpoint VARCHAR(255) NOT NULL,     -- API endpoint
    attempts INTEGER NOT NULL DEFAULT 1,
    window_start TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    blocked_until TIMESTAMP,            -- Optional: block after too many attempts
    
    -- Constraints
    CONSTRAINT unique_rate_limit UNIQUE(identifier, endpoint)
);

CREATE INDEX idx_rate_limits_identifier ON rate_limits(identifier);
CREATE INDEX idx_rate_limits_window_start ON rate_limits(window_start);

-- WebAuthn challenges table (temporary storage)
-- Stores challenges for WebAuthn registration/authentication
CREATE TABLE webauthn_challenges (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    challenge BYTEA NOT NULL,           -- 32-byte random challenge
    challenge_type VARCHAR(20) NOT NULL, -- 'registration' or 'authentication'
    expires_at TIMESTAMP NOT NULL,      -- Challenge expiration (60 seconds)
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT valid_challenge_type CHECK (challenge_type IN ('registration', 'authentication'))
);

CREATE INDEX idx_webauthn_challenges_user_id ON webauthn_challenges(user_id);
CREATE INDEX idx_webauthn_challenges_expires_at ON webauthn_challenges(expires_at);

-- Sessions table
-- JWT session tracking (optional, for refresh tokens)
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id UUID REFERENCES devices(id) ON DELETE CASCADE,
    refresh_token_hash VARCHAR(255) NOT NULL,  -- Hashed refresh token
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_refresh_token_hash ON sessions(refresh_token_hash);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to auto-update updated_at on users table
CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Function to clean up expired challenges (run periodically)
CREATE OR REPLACE FUNCTION cleanup_expired_challenges()
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM webauthn_challenges
    WHERE expires_at < CURRENT_TIMESTAMP;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Function to clean up old audit logs (run periodically, keep last 90 days)
CREATE OR REPLACE FUNCTION cleanup_old_audit_logs(retention_days INTEGER DEFAULT 90)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM audit_logs
    WHERE created_at < CURRENT_TIMESTAMP - (retention_days || ' days')::INTERVAL;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Comments for documentation
COMMENT ON TABLE users IS 'User accounts (zero-knowledge: no plaintext passwords)';
COMMENT ON TABLE vaults IS 'Encrypted vault blobs (server cannot decrypt)';
COMMENT ON TABLE devices IS 'User devices for multi-device sync';
COMMENT ON TABLE passkeys IS 'WebAuthn passkey credentials (public keys only)';
COMMENT ON TABLE audit_logs IS 'Security event audit trail';
COMMENT ON TABLE rate_limits IS 'Rate limiting to prevent brute-force attacks';
COMMENT ON TABLE webauthn_challenges IS 'Temporary WebAuthn challenge storage';
COMMENT ON TABLE sessions IS 'User session management';

COMMENT ON COLUMN users.auk_hash IS 'Bcrypt hash of Account Unlock Key (derived from MEK via HKDF)';
COMMENT ON COLUMN vaults.encrypted_blob IS 'Entire vault encrypted with MEK (zero-knowledge)';
COMMENT ON COLUMN passkeys.credential_public_key IS 'WebAuthn public key (private key never leaves device)';
COMMENT ON COLUMN passkeys.counter IS 'Signature counter to detect credential cloning';
