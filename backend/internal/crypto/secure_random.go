package crypto

import (
	"crypto/rand"
	"io"

	"github.com/rootlock/rootlock/backend/pkg/errors"
)

// GenerateRandomBytes generates cryptographically secure random bytes of the specified length.
// It uses crypto/rand which is backed by the OS's CSPRNG.
//
// Parameters:
//   - length: Number of random bytes to generate
//
// Returns:
//   - []byte: Random bytes
//   - error: Error if random generation fails
func GenerateRandomBytes(length int) ([]byte, error) {
	if length <= 0 {
		return nil, errors.ErrInvalidKeyLength
	}

	randomBytes := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, randomBytes); err != nil {
		return nil, errors.ErrRandomGenerationFailed
	}

	return randomBytes, nil
}

// GenerateNonce generates a cryptographically secure 96-bit (12-byte) nonce for AES-GCM.
// This is the recommended nonce size for AES-GCM to ensure uniqueness and security.
//
// Returns:
//   - []byte: 12-byte random nonce
//   - error: Error if random generation fails
func GenerateNonce() ([]byte, error) {
	return GenerateRandomBytes(NonceSize)
}

// GenerateSalt generates a cryptographically secure salt of the specified length.
// For Argon2id, the recommended salt length is at least 16 bytes.
//
// Parameters:
//   - length: Salt length in bytes (recommended: 16+)
//
// Returns:
//   - []byte: Random salt
//   - error: Error if random generation fails
func GenerateSalt(length int) ([]byte, error) {
	if length < 16 {
		return nil, errors.ErrInvalidSaltLength
	}

	return GenerateRandomBytes(length)
}

// GenerateSecretKey generates a cryptographically secure 256-bit (32-byte) secret key.
// This is used as the user's Secret Key in RootLock's two-secret model.
//
// Returns:
//   - []byte: 32-byte random secret key
//   - error: Error if random generation fails
func GenerateSecretKey() ([]byte, error) {
	return GenerateRandomBytes(32)
}
