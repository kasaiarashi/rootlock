# RootLock Architecture

This document provides a comprehensive overview of RootLock's system architecture, component interactions, and design decisions.

---

## Table of Contents

1. [System Overview](#system-overview)
2. [Component Architecture](#component-architecture)
3. [Security Architecture](#security-architecture)
4. [Data Flow](#data-flow)
5. [Deployment Architecture](#deployment-architecture)
6. [Technology Choices](#technology-choices)

---

## System Overview

RootLock is a **zero-knowledge, end-to-end encrypted password manager** with passkey support, built using modern native technologies across multiple platforms.

### Core Principles

1. **Zero-Knowledge Architecture** - Server cannot decrypt user data
2. **Security-First Design** - Every component prioritizes security
3. **Cross-Platform Native** - Native apps for best performance and UX
4. **Open Source** - Transparent, auditable codebase
5. **Self-Hostable** - Users control their infrastructure

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Client Applications                      │
├──────────┬──────────┬──────────┬──────────┬─────────────────┤
│   iOS    │ Desktop  │ Browser  │   CLI    │   Web Vault     │
│ (Swift)  │ (Tauri)  │ (MV3)    │ (Rust)   │ (React)         │
└──────────┴──────────┴──────────┴──────────┴─────────────────┘
                            ↕ HTTPS (TLS 1.3)
┌─────────────────────────────────────────────────────────────┐
│                      Backend API (Go)                        │
├─────────────────────────────────────────────────────────────┤
│  • Authentication (WebAuthn + Password)                      │
│  • Encrypted Vault Sync                                      │
│  • Device Management                                         │
│  • Audit Logging                                             │
│  • Rate Limiting                                             │
└─────────────────────────────────────────────────────────────┘
                            ↕
┌──────────────────────┬──────────────────────────────────────┐
│  PostgreSQL 16       │         Redis 7                      │
│  • User accounts     │  • Rate limiting                     │
│  • Encrypted vaults  │  • Session storage                   │
│  • Passkeys          │  • Challenge cache                   │
│  • Devices           │  • Temporary data                    │
│  • Audit logs        │                                      │
└──────────────────────┴──────────────────────────────────────┘
```

---

## Component Architecture

### 1. Backend (Go)

**Responsibilities:**
- User authentication (password-based and passkey-based)
- Encrypted vault storage and synchronization
- Device registration and management
- WebAuthn challenge generation and verification
- Rate limiting and security monitoring
- Audit logging

**Technology Stack:**
- **Language**: Go 1.21+
- **Framework**: Gin (HTTP router)
- **ORM**: GORM
- **Database**: PostgreSQL 16
- **Cache**: Redis 7
- **Cryptography**: golang.org/x/crypto (Argon2id, HKDF, AES-GCM)
- **WebAuthn**: github.com/go-webauthn/webauthn

**Directory Structure:**
```
backend/
├── cmd/server/          # Application entry point
├── internal/
│   ├── auth/            # Authentication logic (JWT, sessions)
│   ├── crypto/          # Cryptographic primitives (Argon2id, AES-GCM, HKDF)
│   ├── vault/           # Vault operations (CRUD, sync)
│   ├── user/            # User management
│   ├── device/          # Device registration and trust
│   ├── webauthn/        # Passkey operations (WebAuthn)
│   ├── audit/           # Security audit logging
│   └── sync/            # Synchronization protocol
├── api/
│   ├── handlers/        # HTTP handlers
│   ├── middleware/      # Rate limiting, CORS, security headers
│   └── router.go        # Route definitions
├── db/
│   └── migrations/      # SQL migration scripts
├── config/              # Configuration management
└── pkg/errors/          # Custom error types
```

**Security Considerations:**
- All sensitive operations use constant-time comparisons (`crypto/subtle`)
- Random nonce generation uses `crypto/rand`
- Memory zeroing with `clear()` (Go 1.21+)
- Rate limiting on all authentication endpoints
- Comprehensive audit logging

---

### 2. iOS App (Swift + SwiftUI)

**Responsibilities:**
- Master password management
- Secret Key generation and storage (Keychain)
- Local vault encryption/decryption
- Password AutoFill extension
- Passkey management (ASAuthorizationController)
- Biometric authentication (Face ID/Touch ID)
- Secure Enclave integration

**Technology Stack:**
- **Language**: Swift 5.9+
- **UI Framework**: SwiftUI
- **Cryptography**: CryptoKit (PBKDF2, AES-GCM, HKDF)
- **Secure Storage**: Keychain + Secure Enclave
- **Passkeys**: AuthenticationServices (ASAuthorization)
- **Minimum iOS**: 17.0+

**Directory Structure:**
```
mobile/ios/RootLock/
├── RootLockApp.swift                    # App entry point
├── Models/
│   ├── VaultItem.swift                  # Vault data models
│   ├── User.swift
│   └── Passkey.swift
├── ViewModels/
│   ├── VaultViewModel.swift             # State management
│   ├── AuthViewModel.swift
│   └── PasskeyViewModel.swift
├── Views/
│   ├── Onboarding/                      # First-run experience
│   ├── Vault/                           # Main vault UI
│   ├── Settings/                        # Settings screens
│   ├── Unlock/                          # Biometric unlock
│   └── Passkeys/                        # Passkey management
├── Services/
│   ├── CryptoService.swift              # Encryption operations
│   ├── KeychainService.swift            # Keychain operations
│   ├── SecureEnclaveService.swift       # Secure Enclave
│   ├── PasskeyService.swift             # WebAuthn operations
│   └── SyncService.swift                # Server synchronization
└── RootLockAutofill/                    # Password AutoFill extension
    └── CredentialProviderViewController.swift
```

**Security Considerations:**
- No `String` for sensitive data (use `Data`)
- Memory zeroing via `defer` blocks
- Keychain access: `kSecAttrAccessibleWhenUnlockedThisDeviceOnly`
- File protection: `.completeFileProtection`
- Master key wrapping via Secure Enclave

---

### 3. Desktop App (Tauri + Rust + React)

**Responsibilities:**
- Cross-platform vault access (macOS, Windows, Linux)
- OS keyring integration
- Import from other password managers
- Passkey creation and management
- Browser integration

**Technology Stack:**
- **Backend**: Rust 1.75+
- **Frontend**: React 18 + TypeScript + Tailwind CSS
- **Framework**: Tauri 2.0
- **Cryptography**: argon2, aes-gcm, hkdf, zeroize (Rust crates)
- **OS Integration**: keyring (macOS Keychain, Windows Credential Manager, Linux Secret Service)

**Directory Structure:**
```
desktop/
├── src-tauri/                           # Rust backend
│   ├── src/
│   │   ├── main.rs
│   │   ├── crypto/                      # Cryptographic operations
│   │   ├── vault/                       # Vault storage and sync
│   │   ├── webauthn/                    # Passkey operations
│   │   ├── keyring/                     # OS keyring integration
│   │   └── commands/                    # Tauri commands (IPC)
│   ├── Cargo.toml
│   └── tauri.conf.json
└── ui/                                  # React frontend
    ├── src/
    │   ├── App.tsx
    │   ├── components/                  # React components
    │   ├── hooks/                       # Custom hooks
    │   ├── services/                    # API clients
    │   ├── store/                       # State management (Zustand)
    │   ├── styles/                      # CSS and themes
    │   └── types/                       # TypeScript types
    ├── package.json
    ├── tsconfig.json
    └── tailwind.config.js
```

**Security Considerations:**
- Rust memory safety (no buffer overflows)
- `zeroize` crate for sensitive data cleanup
- Tauri IPC command whitelist
- CSP enforcement
- No sensitive data in localStorage

---

### 4. Browser Extension (Chrome/Edge - Manifest V3)

**Responsibilities:**
- Form detection and credential autofill
- Passkey interception (WebAuthn API)
- Secure communication with desktop app (native messaging)
- Conditional UI (autofill dropdown)

**Technology Stack:**
- **Language**: TypeScript
- **Manifest**: Manifest V3
- **Cryptography**: Web Crypto API
- **Storage**: chrome.storage.session (ephemeral), chrome.storage.local (encrypted only)

**Directory Structure:**
```
browser-extension/
├── shared/                              # Shared code
│   ├── background.ts                    # Service worker
│   ├── content.ts                       # Content script
│   ├── popup.tsx                        # Popup UI
│   ├── messaging/
│   │   ├── nativeHost.ts                # Native messaging protocol
│   │   └── protocol.ts
│   └── autofill/
│       ├── detector.ts                  # Form detection
│       ├── matcher.ts                   # URL matching (eTLD+1)
│       └── injector.ts                  # Credential injection
├── chrome/
│   ├── manifest.json
│   └── native-host/
│       └── com.rootlock.host.json
└── edge/
    └── manifest.json
```

**Security Considerations:**
- Strict CSP (`script-src 'self'`)
- eTLD+1 domain matching (Public Suffix List)
- Homograph attack detection
- Secure credential injection (native value setters)
- HMAC-signed native messaging

---

### 5. CLI Tool (Rust)

**Responsibilities:**
- Vault operations from command line
- Password generation
- Import/export utilities
- Headless environment support

**Technology Stack:**
- **Language**: Rust 1.75+
- **CLI Framework**: clap
- **Async Runtime**: tokio
- **Cryptography**: argon2, aes-gcm, hkdf

**Directory Structure:**
```
cli/
├── src/
│   ├── main.rs
│   ├── commands/                        # CLI commands
│   ├── vault/                           # Vault operations
│   ├── crypto/                          # Cryptographic operations
│   ├── api/                             # HTTP client
│   └── config/                          # Configuration
└── Cargo.toml
```

---

## Security Architecture

### Two-Secret Model

RootLock implements a **dual-secret authentication model** inspired by 1Password:

```
┌──────────────────────────────────────────────────────────┐
│ User Secrets (Never Sent to Server)                      │
├──────────────────────────────────────────────────────────┤
│ 1. Master Password (user-chosen, high-entropy)           │
│ 2. Secret Key (256-bit random, generated client-side)    │
└──────────────────────────────────────────────────────────┘
                         ↓
┌──────────────────────────────────────────────────────────┐
│ Key Derivation (Argon2id)                                │
├──────────────────────────────────────────────────────────┤
│ Argon2id(                                                │
│   password: Master Password,                             │
│   salt: User UUID || Secret Key,                         │
│   memory: 64 MB,                                         │
│   iterations: 3,                                         │
│   parallelism: 4,                                        │
│   output: 256 bits                                       │
│ )                                                        │
└──────────────────────────────────────────────────────────┘
                         ↓
┌──────────────────────────────────────────────────────────┐
│ Master Encryption Key (MEK)                              │
│ • Used to derive all item encryption keys                │
│ • Never sent to server                                   │
│ • Never stored in plaintext                              │
└──────────────────────────────────────────────────────────┘
                         ↓
          ┌──────────────┴──────────────┐
          ↓                             ↓
┌─────────────────────┐    ┌────────────────────────┐
│ Account Unlock Key  │    │ Item Encryption Keys   │
│ (AUK)               │    │ (IEK)                  │
├─────────────────────┤    ├────────────────────────┤
│ HKDF(MEK,           │    │ HKDF(MEK,              │
│   info="account")   │    │   salt=item_uuid,      │
│                     │    │   info="vault-item")   │
│ → Sent to server    │    │                        │
│   (bcrypt hashed)   │    │ → Encrypt vault items  │
└─────────────────────┘    └────────────────────────┘
```

### Encryption Hierarchy

**Per-Item Encryption:**
```
Each vault item has:
1. Unique encryption key (IEK) derived via HKDF from MEK
2. Unique nonce (96-bit random, never reused)
3. AES-256-GCM encryption with authentication tag
4. Additional Authenticated Data (AAD): item_uuid || item_type || version
```

**Encrypted Item Structure:**
```json
{
  "id": "uuid-v4",
  "user_id": "uuid-v4",
  "encrypted_data": "base64(nonce || ciphertext || tag)",
  "item_type": "password",  // Not encrypted (for filtering)
  "version": 1,
  "created_at": "2026-01-16T...",
  "updated_at": "2026-01-16T..."
}
```

### Passkey Architecture

**WebAuthn/FIDO2 Integration:**

```
┌──────────────────────────────────────────────────────────┐
│ Passkey Registration (Website)                           │
├──────────────────────────────────────────────────────────┤
│ 1. User visits website, clicks "Create passkey"          │
│ 2. Website calls navigator.credentials.create()          │
│ 3. Browser shows passkey creation prompt                 │
│ 4. User authenticates (Face ID/Touch ID/Windows Hello)   │
│ 5. Device generates key pair (never exports private key) │
│ 6. Public key sent to website, stored in database        │
│ 7. RootLock intercepts and saves passkey metadata        │
└──────────────────────────────────────────────────────────┘
                         ↓
┌──────────────────────────────────────────────────────────┐
│ Passkey Authentication (Website)                         │
├──────────────────────────────────────────────────────────┤
│ 1. User visits website, clicks "Sign in with passkey"    │
│ 2. Website calls navigator.credentials.get()             │
│ 3. Browser/RootLock shows passkey selection              │
│ 4. User selects passkey and authenticates                │
│ 5. Device signs challenge with private key               │
│ 6. Signature sent to website, verified with public key   │
│ 7. User signed in (no password needed)                   │
└──────────────────────────────────────────────────────────┘
```

**Passkey Storage:**
- **iOS**: Stored in iCloud Keychain (synced across Apple devices)
- **Desktop**: Stored in OS keyring (platform-specific)
- **Browser**: RootLock extension intercepts and stores metadata
- **Backend**: Stores public keys only (for account sign-in with passkeys)

---

## Data Flow

### Account Creation Flow

```
┌─────────┐                ┌─────────┐                ┌─────────┐
│ Client  │                │ Backend │                │Database │
└────┬────┘                └────┬────┘                └────┬────┘
     │                          │                          │
     │ 1. User enters email     │                          │
     │    and master password   │                          │
     │                          │                          │
     │ 2. Generate Secret Key   │                          │
     │    (256-bit random)      │                          │
     │                          │                          │
     │ 3. Derive MEK via        │                          │
     │    Argon2id              │                          │
     │                          │                          │
     │ 4. Derive AUK via HKDF   │                          │
     │                          │                          │
     │ 5. POST /api/v1/auth/register                       │
     │    (email, AUK hash) ───────────>                   │
     │                          │                          │
     │                          │ 6. Create user record    │
     │                          │    (email, AUK hash) ───>│
     │                          │                          │
     │                          │<──── User created ───────│
     │                          │                          │
     │<─── 200 OK (user_id) ────│                          │
     │                          │                          │
     │ 7. Store Secret Key in   │                          │
     │    secure storage        │                          │
     │    (Keychain/OS Keyring) │                          │
     │                          │                          │
     │ 8. Show Secret Key to    │                          │
     │    user (QR + PDF backup)│                          │
     │                          │                          │
```

### Vault Sync Flow

```
┌─────────┐                ┌─────────┐                ┌─────────┐
│ Client  │                │ Backend │                │Database │
└────┬────┘                └────┬────┘                └────┬────┘
     │                          │                          │
     │ 1. User unlocks vault    │                          │
     │    (derives MEK locally) │                          │
     │                          │                          │
     │ 2. GET /api/v1/vault ────────────>                  │
     │    (Authorization: Bearer JWT)                      │
     │                          │                          │
     │                          │ 3. Fetch encrypted vault │
     │                          │<─────────────────────────│
     │                          │                          │
     │<── Encrypted vault blob ─│                          │
     │                          │                          │
     │ 4. Decrypt vault locally │                          │
     │    using MEK             │                          │
     │                          │                          │
     │ 5. User adds/edits item  │                          │
     │                          │                          │
     │ 6. Encrypt item with IEK │                          │
     │    (derived from MEK)    │                          │
     │                          │                          │
     │ 7. PUT /api/v1/vault ─────────────>                 │
     │    (encrypted blob)      │                          │
     │                          │                          │
     │                          │ 8. Store encrypted blob  │
     │                          │  ───────────────────────>│
     │                          │                          │
     │                          │<──── Success ────────────│
     │                          │                          │
     │<──── 200 OK ─────────────│                          │
     │                          │                          │
```

### Passkey Authentication Flow (Sign In to RootLock)

```
┌─────────┐                ┌─────────┐                ┌─────────┐
│ Client  │                │ Backend │                │Database │
└────┬────┘                └────┬────┘                └────┬────┘
     │                          │                          │
     │ 1. POST /api/v1/webauthn/authenticate/options       │
     │ ────────────────────────>│                          │
     │                          │                          │
     │                          │ 2. Generate challenge    │
     │                          │    (32-byte random)      │
     │                          │                          │
     │                          │ 3. Fetch user's passkeys │
     │                          │<─────────────────────────│
     │                          │                          │
     │<─ Challenge + options ───│                          │
     │                          │                          │
     │ 4. navigator.credentials │                          │
     │    .get()                │                          │
     │                          │                          │
     │ 5. User authenticates    │                          │
     │    (biometric)           │                          │
     │                          │                          │
     │ 6. Device signs challenge│                          │
     │    with private key      │                          │
     │                          │                          │
     │ 7. POST /api/v1/webauthn/authenticate/verify        │
     │    (credential + signature)                         │
     │ ────────────────────────>│                          │
     │                          │                          │
     │                          │ 8. Verify signature with │
     │                          │    stored public key     │
     │                          │<─────────────────────────│
     │                          │                          │
     │                          │ 9. Create JWT session    │
     │                          │                          │
     │<──── JWT token ──────────│                          │
     │                          │                          │
     │ 10. User signed in       │                          │
     │                          │                          │
```

---

## Deployment Architecture

### Self-Hosted (Docker Compose)

**Simple single-server deployment:**

```
┌───────────────────────────────────────────────────────────┐
│                    Docker Host                             │
├───────────────────────────────────────────────────────────┤
│                                                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │   Caddy      │  │  RootLock    │  │  PostgreSQL  │    │
│  │  (Reverse    │  │   Backend    │  │      16      │    │
│  │   Proxy)     │  │    (Go)      │  │              │    │
│  │              │  │              │  │              │    │
│  │ Auto HTTPS   │  │ Port 8080    │  │ Port 5432    │    │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘    │
│         │                 │                 │             │
│         │                 │  ┌──────────────┘             │
│         │                 │  │  ┌──────────────┐          │
│         │                 │  │  │    Redis     │          │
│         │                 │  │  │      7       │          │
│         │                 │  │  │              │          │
│         │                 │  │  │ Port 6379    │          │
│         │                 │  │  └──────────────┘          │
│         │                 │  │                            │
│         └─────────────────┴──┴────────────────────────    │
│                                                            │
└───────────────────────────────────────────────────────────┘
                       ↑ HTTPS (443)
                       │
                 Internet Users
```

**docker-compose.yml:**
```yaml
version: '3.9'

services:
  backend:
    build: ./backend
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgresql://rootlock:${DB_PASSWORD}@postgres:5432/rootlock
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:16-alpine
    volumes:
      - vault_data:/var/lib/postgresql/data
    environment:
      - POSTGRES_DB=rootlock
      - POSTGRES_USER=rootlock
      - POSTGRES_PASSWORD=${DB_PASSWORD}

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

  caddy:
    image: caddy:2-alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
      - caddy_data:/data
      - caddy_config:/config

volumes:
  vault_data:
  redis_data:
  caddy_data:
  caddy_config:
```

### Cloud-Native (Kubernetes)

**Scalable multi-region deployment:**

```
┌──────────────────────────────────────────────────────────┐
│                 Kubernetes Cluster                        │
├──────────────────────────────────────────────────────────┤
│                                                           │
│  ┌─────────────────────────────────────────────────┐    │
│  │  Ingress (NGINX or Traefik)                     │    │
│  │  • TLS termination                              │    │
│  │  • Load balancing                               │    │
│  └────────────────────┬────────────────────────────┘    │
│                       │                                  │
│  ┌────────────────────┴────────────────────────────┐    │
│  │  RootLock Backend Pods (HPA: 2-10 replicas)     │    │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐      │    │
│  │  │ Backend  │  │ Backend  │  │ Backend  │      │    │
│  │  │  Pod 1   │  │  Pod 2   │  │  Pod N   │      │    │
│  │  └──────────┘  └──────────┘  └──────────┘      │    │
│  └─────────────────────────────────────────────────┘    │
│                       │                                  │
│         ┌─────────────┴─────────────┐                   │
│         │                           │                   │
│  ┌──────┴──────────┐   ┌───────────┴────────┐          │
│  │  PostgreSQL     │   │  Redis Cluster      │          │
│  │  StatefulSet    │   │  (Master/Replica)   │          │
│  │  (Primary +     │   │                     │          │
│  │   Replicas)     │   │  ┌───────┐          │          │
│  │                 │   │  │Master │          │          │
│  │  Persistent     │   │  └───┬───┘          │          │
│  │  Volumes        │   │      │              │          │
│  │                 │   │  ┌───┴───┐          │          │
│  └─────────────────┘   │  │Replica│          │          │
│                        │  └───────┘          │          │
│                        └─────────────────────┘          │
│                                                          │
└──────────────────────────────────────────────────────────┘
```

**Key Kubernetes Resources:**
- **Deployment**: RootLock backend (stateless)
- **StatefulSet**: PostgreSQL (stateful)
- **Service**: ClusterIP for internal communication
- **Ingress**: HTTPS termination and routing
- **HPA**: Horizontal Pod Autoscaler (scale 2-10 pods)
- **PVC**: Persistent Volume Claims for database
- **Secret**: Database credentials, JWT secrets
- **ConfigMap**: Application configuration

---

## Technology Choices

### Why Go for Backend?

1. **Performance**: Compiled, statically-typed, fast execution
2. **Concurrency**: Goroutines for handling many connections
3. **Standard Library**: Excellent crypto libraries (`crypto/*`, `golang.org/x/crypto`)
4. **Deployment**: Single binary, easy to containerize
5. **Ecosystem**: Rich ecosystem for web services (Gin, GORM)
6. **Security**: Memory safety (garbage collected), no buffer overflows

### Why Swift/SwiftUI for iOS?

1. **Native Performance**: Direct access to iOS frameworks
2. **CryptoKit**: Apple's modern cryptography framework
3. **Secure Enclave**: Hardware-backed key storage
4. **AutoFill**: Native integration with iOS Password AutoFill
5. **SwiftUI**: Declarative UI, modern development experience
6. **Type Safety**: Prevents many common bugs

### Why Tauri for Desktop?

1. **Security**: Rust backend (memory safety, no vulnerabilities)
2. **Small Footprint**: ~3-5 MB binaries (vs Electron's ~100+ MB)
3. **Native Performance**: Uses system WebView (no bundled Chromium)
4. **Cross-Platform**: One codebase for macOS, Windows, Linux
5. **Modern Stack**: React + TypeScript for familiar frontend development
6. **Active Community**: Growing ecosystem, good documentation

### Why Rust for CLI?

1. **Performance**: Compiled, zero-cost abstractions
2. **Memory Safety**: No segfaults, no buffer overflows
3. **Cross-Platform**: Single codebase compiles to all platforms
4. **Excellent Crypto Crates**: RustCrypto (argon2, aes-gcm, etc.)
5. **Binary Distribution**: Single executable, no runtime dependencies
6. **Correctness**: Type system prevents many bugs

### Why PostgreSQL?

1. **ACID Compliance**: Strong transactional guarantees
2. **JSONB Support**: Efficient storage for encrypted vault blobs
3. **Reliability**: Battle-tested, widely deployed
4. **Extensions**: pgcrypto for additional crypto operations
5. **Replication**: Built-in support for backups and scaling
6. **Open Source**: Free, auditable, community-supported

### Why Redis?

1. **Speed**: In-memory performance for session storage
2. **TTL Support**: Automatic expiration for challenges
3. **Atomic Operations**: Safe rate limiting implementation
4. **Persistence Options**: RDB + AOF for durability
5. **Simple**: Easy to deploy and maintain
6. **Scalable**: Redis Cluster for high availability

---

## Conclusion

RootLock's architecture prioritizes **security, performance, and user experience** across all platforms. The zero-knowledge design ensures that even with full database access, an attacker cannot decrypt user data. The use of modern, native technologies provides excellent performance while maintaining a small footprint.

The architecture is designed to be **auditable, extensible, and maintainable**, with clear separation of concerns and comprehensive documentation.

---

**Next Steps:**
- Review [Security Documentation](security.md) for threat model
- Read [Crypto Specification](crypto-spec.md) for implementation details
- See [API Documentation](api.md) for backend endpoints
- Check [Development Guide](development.md) for setup instructions
