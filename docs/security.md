# RootLock Security Model

## Table of Contents

1. [Threat Model](#threat-model)
2. [Security Architecture](#security-architecture)
3. [Cryptographic Guarantees](#cryptographic-guarantees)
4. [Attack Surface Analysis](#attack-surface-analysis)
5. [Security Audit Checklist](#security-audit-checklist)

---

## Threat Model

### Adversaries

1. **Network Attacker** - Can intercept network traffic
2. **Malicious Server Operator** - Has full database access
3. **Compromised Backend** - Attacker gains backend server access
4. **Local Malware** - Runs on user's device
5. **Phishing Attacker** - Attempts to steal credentials

### Security Goals

✅ **Confidentiality**: Vault data unreadable without Master Password + Secret Key  
✅ **Integrity**: Tampering detected via authenticated encryption  
✅ **Availability**: Sync enables recovery from device loss  
✅ **Authentication**: WebAuthn prevents credential phishing  
✅ **Forward Secrecy**: Compromise doesn't reveal past data  

### Non-Goals

❌ **Malicious Client**: User's device with malware (out of scope)  
❌ **Side-Channel Attacks**: Timing/power analysis (mitigated but not eliminated)  
❌ **Quantum Resistance**: AES-256 provides 128-bit quantum security (adequate)  

---

## Security Architecture

### Zero-Knowledge Guarantee

**Server Cannot Decrypt**:
- Master Password never sent to server
- Secret Key never sent to server
- MEK derived client-side only
- Server receives only AUK hash (one-way)

**Verification**:
```go
// Server verification (backend/internal/auth/service.go)
func (s *AuthService) VerifyAUK(providedAUK string, storedHash string) bool {
    // Constant-time comparison
    err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(providedAUK))
    return err == nil
}
```

### Cryptographic Primitives

| Operation | Algorithm | Parameters |
|-----------|-----------|-----------|
| Key Derivation | Argon2id | m=64MB, t=3, p=4 |
| Encryption | AES-256-GCM | 256-bit key, 96-bit nonce |
| Key Expansion | HKDF-SHA256 | Per-item salts |
| Randomness | crypto/rand | OS entropy source |
| Hashing | SHA-256 | For integrity checks |
| Password Hashing | bcrypt | Cost factor 12 |

**Security Properties**:
- **Argon2id**: Memory-hard (64MB), resistant to GPU/ASIC attacks
- **AES-256-GCM**: Authenticated encryption, detects tampering
- **HKDF**: Cryptographically strong key derivation
- **Bcrypt**: Slow hashing, resistant to brute force

---

## Attack Surface Analysis

### 1. Network Attacks

**Threat**: Man-in-the-middle, eavesdropping

**Mitigations**:
- ✅ TLS 1.3 enforcement
- ✅ Certificate pinning (mobile apps)
- ✅ HSTS headers
- ✅ Only encrypted blobs transmitted
- ✅ No sensitive data in URLs/headers

### 2. Server Compromise

**Threat**: Attacker gains database access

**Mitigations**:
- ✅ Zero-knowledge architecture (cannot decrypt)
- ✅ AUK is hashed (cannot reverse)
- ✅ Audit logs detect suspicious activity
- ✅ Rate limiting prevents brute force

**Impact if compromised**:
- ❌ Can see user emails, metadata
- ❌ Can see encrypted vault blobs
- ✅ Cannot decrypt vaults (no MEK)
- ✅ Cannot impersonate users (no passwords)

### 3. Client-Side Attacks

**Threat**: Malware on user device

**Mitigations**:
- ✅ Memory zeroing after use
- ✅ Secure storage (Keychain/OS Keyring)
- ✅ Auto-lock timeouts
- ✅ Clipboard auto-clear
- ✅ Screen capture protection (where possible)

**Residual Risk**: Malware with root access can steal from secure storage

### 4. Phishing Attacks

**Threat**: Fake website steals credentials

**Mitigations (Passkeys)**:
- ✅ Passkeys bound to origin (cannot be phished)
- ✅ Browser enforces domain matching
- ✅ No secrets to phish

**Mitigations (Passwords)**:
- ✅ eTLD+1 domain matching (browser extension)
- ✅ Homograph attack detection
- ✅ Typosquatting warnings

### 5. Brute Force Attacks

**Threat**: Offline password cracking

**Mitigations**:
- ✅ Argon2id (expensive to compute)
- ✅ Two secrets required (Master Password + Secret Key)
- ✅ High entropy requirements

**Attack Cost**: With Argon2id (64MB, 3 iterations):
- ~50ms per attempt on modern CPU
- ~20 attempts/second
- 256-bit Secret Key adds 2^256 keyspace

### 6. WebAuthn-Specific Attacks

**Threat**: Credential replay, phishing

**Mitigations**:
- ✅ Challenge-response (prevents replay)
- ✅ Origin validation (prevents phishing)
- ✅ Counter verification (detects cloning)
- ✅ User verification required
- ✅ Challenge expiration (60 seconds)

---

## Security Audit Checklist

### Cryptography
- [ ] Argon2id parameters meet OWASP recommendations
- [ ] AES-GCM nonces never reused
- [ ] HKDF info strings unique per context
- [ ] Random number generation uses crypto/rand
- [ ] Constant-time comparisons for secrets
- [ ] Memory zeroing implemented
- [ ] No hardcoded keys/secrets

### Authentication
- [ ] Rate limiting on login endpoints
- [ ] Account lockout after failed attempts
- [ ] JWT expiration reasonable (15min)
- [ ] Refresh token rotation
- [ ] WebAuthn challenge uniqueness
- [ ] Challenge expiration enforced
- [ ] User verification required

### Storage
- [ ] Database encryption at rest
- [ ] Secure deletion of sensitive data
- [ ] Keychain/OS Keyring integration
- [ ] File permissions restrictive
- [ ] No plaintext secrets in logs

### Network
- [ ] TLS 1.3 enforced
- [ ] Certificate validation
- [ ] HSTS enabled
- [ ] CSP headers configured
- [ ] CORS whitelist only trusted origins

### Input Validation
- [ ] All user inputs sanitized
- [ ] SQL injection prevention (parameterized queries)
- [ ] XSS prevention (CSP + escaping)
- [ ] CSRF tokens on state-changing operations
- [ ] URL validation (eTLD+1 matching)

### Audit & Monitoring
- [ ] Security events logged
- [ ] Anomaly detection configured
- [ ] Failed login attempts tracked
- [ ] Data export events logged
- [ ] Admin actions audited

---

## Security Contact

Report security vulnerabilities to: **security@rootlock.kriaa.in**

**Response Time**: Within 48 hours for critical issues

**Disclosure Policy**: Coordinated disclosure (90 days)

---

## Security Advisories

Check [GitHub Security Advisories](https://github.com/kasaiarashi/rootlock/security/advisories) for published vulnerabilities.

---

**Last Updated**: 2026-01-16  
**Version**: 1.0
