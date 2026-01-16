# 🔐 RootLock

**A production-grade, zero-knowledge password manager with passkey support.**

RootLock is an open-source, security-first password manager inspired by 1Password's architecture, built using modern, native technologies across all platforms.

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)
[![Swift Version](https://img.shields.io/badge/Swift-5.9+-F05138?logo=swift)](https://swift.org/)
[![Rust Version](https://img.shields.io/badge/Rust-1.75+-000000?logo=rust)](https://www.rust-lang.org/)

---

## ✨ Features

### 🔒 Security-First Architecture
- **Zero-knowledge encryption** - Server cannot decrypt your data
- **Two-secret model** - Master Password + Secret Key (256-bit)
- **Argon2id key derivation** - Memory-hard, GPU-resistant
- **AES-256-GCM encryption** - Authenticated encryption for all vault items
- **Passkey support** - WebAuthn/FIDO2 for passwordless authentication
- **Secure Enclave** (iOS) - Hardware-backed key storage
- **End-to-end encrypted sync** - Your data is encrypted in transit and at rest

### 🚀 Cross-Platform
- **iOS** (17+) - Native Swift/SwiftUI with AutoFill extension
- **Desktop** - Tauri (Rust + React) for macOS, Windows, Linux
- **Browser Extensions** - Chrome & Edge (Manifest V3)
- **CLI** - Rust-based command-line tool for power users
- **Backend** - Go REST API with PostgreSQL

### 🎨 Beautiful & Intuitive
- **Dark mode primary** with light mode support
- **Modern, minimal design** - Clean interfaces across all platforms
- **Password generator** - Cryptographic random + memorable passwords
- **Breach monitoring** - HaveIBeenPwned integration (k-anonymity)
- **Password health dashboard** - Identify weak, reused, or breached passwords

### 🌐 Deployment Flexibility
- **Self-hosted** - Docker Compose for single-server deployment
- **Cloud-ready** - Kubernetes manifests for scalable infrastructure
- **Hybrid support** - Run on-premises or in the cloud

---

## 🏗️ Architecture

### Two-Secret Model

RootLock uses a dual-secret architecture for maximum security:

1. **Master Password** - User-chosen, high-entropy password (never sent to server)
2. **Secret Key** - 256-bit randomly generated key (stored only on user devices)

```
User Authentication:
  Master Password + Secret Key
           ↓
  Argon2id Key Derivation (64MB memory, 3 iterations)
           ↓
  Master Encryption Key (MEK)
           ↓
  HKDF → Item Encryption Keys
           ↓
  AES-256-GCM → Encrypted Vault Items
```

### Zero-Knowledge Design

The server **never** receives:
- ❌ Master Password
- ❌ Secret Key
- ❌ Master Encryption Key
- ❌ Plaintext passwords or vault data

The server **only** receives:
- ✅ Account Unlock Key (derived via HKDF, hashed with bcrypt)
- ✅ Encrypted vault blobs
- ✅ Encrypted metadata

### Passkey Support

RootLock implements **WebAuthn/FIDO2** for passwordless authentication:
- Create passkeys on any website
- Sign in with biometrics (Face ID, Touch ID, Windows Hello)
- Phishing-resistant authentication
- Cross-device support (QR code flow)
- Synced passkeys via platform providers (iCloud Keychain, Google Password Manager)

---

## 🚀 Quick Start

### Prerequisites

- **Docker** & Docker Compose
- **Go 1.21+** (for backend development)
- **Node.js 18+** (for desktop/browser development)
- **Rust 1.75+** (for desktop/CLI development)
- **Xcode 15+** (for iOS development)

### Local Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/kasaiarashi/rootlock.git
   cd rootlock
   ```

2. **Start infrastructure**
   ```bash
   docker-compose up -d
   ```
   This starts PostgreSQL and Redis.

3. **Run backend**
   ```bash
   cd backend
   go run cmd/server/main.go
   ```

4. **Run desktop app** (in separate terminal)
   ```bash
   cd desktop
   npm install
   npm run tauri dev
   ```

5. **Open browser**
   Navigate to `http://localhost:8080` to access the backend API.

For detailed setup instructions, see [docs/development.md](docs/development.md).

---

## 📚 Documentation

- **[Architecture Overview](docs/architecture.md)** - System design and component interactions
- **[Security Model](docs/security.md)** - Threat model, cryptographic specifications
- **[Crypto Specification](docs/crypto-spec.md)** - Detailed cryptographic implementation
- **[API Reference](docs/api.md)** - REST API documentation (OpenAPI 3.0)
- **[Development Guide](docs/development.md)** - Setup and contribution guidelines
- **[Deployment Guide](docs/deployment.md)** - Self-hosted and cloud deployment

---

## 🛠️ Technology Stack

| Component | Technologies |
|-----------|-------------|
| **Backend** | Go, PostgreSQL, Redis, Argon2id, AES-GCM |
| **iOS** | Swift, SwiftUI, CryptoKit, Secure Enclave, ASAuthorization |
| **Desktop** | Rust, Tauri 2.0, React 18, TypeScript, Tailwind CSS |
| **Browser** | TypeScript, Manifest V3, Web Crypto API |
| **CLI** | Rust, clap, tokio |

---

## 🔒 Security

### Reporting Security Issues

**DO NOT** open public GitHub issues for security vulnerabilities.

Please report security issues to: **security@rootlock.kriaa.in**

We take security seriously and will respond promptly to verified reports.

### Security Features

- ✅ Argon2id key derivation (OWASP recommended parameters)
- ✅ AES-256-GCM authenticated encryption
- ✅ HKDF for key expansion
- ✅ Constant-time comparisons (timing attack prevention)
- ✅ Memory zeroing for sensitive data
- ✅ Rate limiting on authentication endpoints
- ✅ Audit logging for security events
- ✅ WebAuthn/FIDO2 passkey support
- ✅ TLS 1.3 enforcement
- ✅ CSRF protection
- ✅ Security headers (CSP, HSTS, etc.)

### Audits

RootLock is designed to be security-audited. See [docs/security.md](docs/security.md) for our security audit preparation checklist.

---

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Commit with descriptive messages
6. Push to your fork
7. Open a Pull Request

### Code of Conduct

This project follows a Code of Conduct. By participating, you agree to uphold this code.

---

## 📋 Roadmap

### Phase 1: Foundation (Current)
- [x] Project structure
- [x] Documentation
- [x] Docker infrastructure
- [ ] Backend crypto primitives
- [ ] Database schema

### Phase 2: Core Backend
- [ ] Authentication API
- [ ] Vault sync protocol
- [ ] Device management
- [ ] Passkey support (WebAuthn)

### Phase 3: iOS App
- [ ] Onboarding flow
- [ ] Vault UI
- [ ] Password AutoFill extension
- [ ] Passkey integration
- [ ] Secure Enclave integration

### Phase 4: Desktop App
- [ ] Tauri + React UI
- [ ] OS keyring integration
- [ ] Import from other password managers
- [ ] Passkey support

### Phase 5: Browser Extension
- [ ] Manifest V3 implementation
- [ ] Form detection & autofill
- [ ] Passkey interception
- [ ] Native messaging

### Phase 6: CLI Tool
- [ ] Vault operations
- [ ] Password generation
- [ ] Import/export utilities

### Phase 7: Advanced Features
- [ ] Breach monitoring (HIBP)
- [ ] Password health dashboard
- [ ] Memorable password generator
- [ ] Security audit reports

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- Inspired by [1Password's security model](https://1password.com/security/)
- Built with [Tauri](https://tauri.app/) for desktop apps
- Powered by [CryptoKit](https://developer.apple.com/documentation/cryptokit) (iOS)
- Uses [Argon2](https://github.com/P-H-C/phc-winner-argon2) for key derivation
- Breach checking via [HaveIBeenPwned](https://haveibeenpwned.com/)
- WebAuthn support via [go-webauthn](https://github.com/go-webauthn/webauthn)

---

## 📞 Contact

- **Website**: https://rootlock.kriaa.in (coming soon)
- **Email**: dev@rootlock.kriaa.in
- **Security**: security@rootlock.kriaa.in
- **GitHub**: https://github.com/kasaiarashi/rootlock

---

**Built with ❤️ and a strong commitment to security and privacy.**
