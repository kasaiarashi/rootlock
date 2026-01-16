# Cryptographic Specification

## Overview

RootLock uses industry-standard cryptographic primitives with parameters recommended by OWASP, NIST, and FIDO Alliance.

---

## Key Derivation

### Argon2id Parameters

```
Algorithm: Argon2id (RFC 9106)
Version: 0x13
Memory Cost: 65536 KiB (64 MB)
Time Cost: 3 iterations
Parallelism: 4 threads
Salt: User UUID (16 bytes) || Secret Key (32 bytes) = 48 bytes
Output Length: 32 bytes (256 bits)
```

**Rationale**: OWASP recommendation for password hashing. Memory-hard to resist GPU/ASIC attacks.

**Implementation**:
```go
// Go (backend)
import "golang.org/x/crypto/argon2"

mek := argon2.IDKey(
    []byte(masterPassword),
    append(userUUID, secretKey...),
    3,      // iterations
    65536,  // memory (64 MB)
    4,      // parallelism
    32      // output length
)
```

```swift
// iOS (CryptoKit uses PBKDF2 instead)
import CryptoKit

let derivedKey = try PBKDF2.deriveKey(
    password: masterPassword,
    salt: salt,
    iterations: 600_000, // OWASP 2023 recommendation
    keyLength: 32
)
```

---

## Symmetric Encryption

### AES-256-GCM

```
Algorithm: AES-GCM (NIST SP 800-38D)
Key Size: 256 bits
Nonce Size: 96 bits (12 bytes)
Tag Size: 128 bits (16 bytes)
Mode: Galois/Counter Mode (GCM)
```

**Encryption**:
```
Input: Plaintext, Key (256-bit), AAD (Additional Authenticated Data)
1. Generate nonce (96-bit random via crypto/rand)
2. Encrypt: AES-256-GCM(plaintext, key, nonce, AAD)
3. Output: nonce || ciphertext || tag
```

**Decryption**:
```
Input: nonce || ciphertext || tag, Key, AAD
1. Extract nonce (first 12 bytes)
2. Extract tag (last 16 bytes)
3. Decrypt: AES-256-GCM-Decrypt(ciphertext, key, nonce, AAD, tag)
4. Verify authentication tag
5. Output: plaintext or error
```

**Security Properties**:
- Authenticated encryption (integrity + confidentiality)
- Nonce uniqueness critical (never reuse with same key)
- Tag verification prevents tampering

---

## Key Hierarchy

### Master Encryption Key (MEK)

```
MEK = Argon2id(MasterPassword, salt=UUID||SecretKey)
```

### Account Unlock Key (AUK)

```
AUK = HKDF-SHA256(
    ikm: MEK,
    salt: nil,
    info: "account-unlock-key",
    length: 32
)

Server stores: bcrypt(AUK, cost=12)
```

### Item Encryption Key (IEK)

```
For each vault item:

IEK = HKDF-SHA256(
    ikm: MEK,
    salt: item_uuid (16 bytes),
    info: "vault-item-encryption",
    length: 32
)
```

---

## Passkey Cryptography (WebAuthn)

### Supported Algorithms

```
ES256:  ECDSA with P-256 and SHA-256 (COSE -7)
ES384:  ECDSA with P-384 and SHA-384 (COSE -35)
RS256:  RSA-PSS with SHA-256 (COSE -257)
EdDSA:  Ed25519 (COSE -8)
```

**Preference Order**: EdDSA > ES256 > ES384 > RS256

### Registration Flow

```
1. Server generates challenge (32 bytes random)
2. Client creates key pair (private key never exported)
3. Authenticator signs (authData || clientDataHash) with private key
4. Client sends: public key + signature + attestation
5. Server verifies signature, stores public key
```

### Authentication Flow

```
1. Server generates challenge (32 bytes random)
2. Client signs (authData || clientDataHash) with private key
3. Client sends: credentialID + signature + authenticatorData
4. Server verifies signature with stored public key
5. Server checks counter (prevent cloning)
```

---

## Random Number Generation

All random values use cryptographically secure sources:

- **Go**: `crypto/rand`
- **Swift**: `SystemRandomNumberGenerator`
- **Rust**: `rand::thread_rng()` (backed by `getrandom`)
- **Browser**: `crypto.getRandomValues()`

**Never use**: `math/rand`, `Math.random()`, or predictable sources.

---

## Constant-Time Operations

All secret comparisons use constant-time functions to prevent timing attacks:

```go
// Go
import "crypto/subtle"
if subtle.ConstantTimeCompare(expected, provided) == 1 {
    // Equal
}
```

```rust
// Rust
use subtle::ConstantTimeEq;
if expected.ct_eq(&provided).into() {
    // Equal
}
```

---

## Memory Safety

### Zeroing Sensitive Data

```go
// Go
defer clear(sensitiveSlice)
```

```rust
// Rust
use zeroize::Zeroize;
sensitive_data.zeroize();
```

```swift
// Swift
defer {
    sensitiveData.resetBytes(in: 0..<sensitiveData.count)
}
```

---

## Test Vectors

### Argon2id

```
Password: "test-password"
Salt: hex"0123456789abcdef0123456789abcdef"
Output (hex): [implementation-specific]
```

### AES-256-GCM

```
Key: hex"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
Nonce: hex"0123456789abcdef01234567"
Plaintext: "Hello, RootLock!"
AAD: hex"01020304"
Ciphertext: [implementation-specific]
Tag: [implementation-specific]
```

---

## References

- [RFC 9106 - Argon2](https://datatracker.ietf.org/doc/html/rfc9106)
- [NIST SP 800-38D - GCM](https://nvlpubs.nist.gov/nistpubs/Legacy/SP/nistspecialpublication800-38d.pdf)
- [RFC 5869 - HKDF](https://datatracker.ietf.org/doc/html/rfc5869)
- [WebAuthn Level 3](https://www.w3.org/TR/webauthn-3/)
- [OWASP Password Storage](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)

---

**Version**: 1.0  
**Last Updated**: 2026-01-16
