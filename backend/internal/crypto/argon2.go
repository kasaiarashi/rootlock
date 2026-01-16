package crypto

import (
	"golang.org/x/crypto/argon2"

	"github.com/rootlock/rootlock/backend/pkg/errors"
)

// Argon2 parameters following OWASP recommendations for password hashing
const (
	// Argon2Memory is the memory cost in KiB (64 MB)
	Argon2Memory uint32 = 64 * 1024

	// Argon2Iterations is the number of iterations (time cost)
	Argon2Iterations uint32 = 3

	// Argon2Parallelism is the number of threads to use
	Argon2Parallelism uint8 = 4

	// Argon2KeyLength is the output key length in bytes (256 bits)
	Argon2KeyLength uint32 = 32

	// MinSaltLength is the minimum recommended salt length
	MinSaltLength = 16
)

// Argon2Params contains the parameters for Argon2id key derivation.
type Argon2Params struct {
	Memory      uint32 // Memory cost in KiB
	Iterations  uint32 // Time cost (number of iterations)
	Parallelism uint8  // Number of threads
	SaltLength  uint32 // Salt length in bytes
	KeyLength   uint32 // Output key length in bytes
}

// DefaultArgon2Params returns the default Argon2id parameters recommended by OWASP.
// These parameters provide strong resistance against brute-force attacks while
// maintaining reasonable performance on modern hardware.
//
// Returns:
//   - *Argon2Params: Default parameters (64MB memory, 3 iterations, 4 threads)
func DefaultArgon2Params() *Argon2Params {
	return &Argon2Params{
		Memory:      Argon2Memory,
		Iterations:  Argon2Iterations,
		Parallelism: Argon2Parallelism,
		SaltLength:  MinSaltLength,
		KeyLength:   Argon2KeyLength,
	}
}

// DeriveKey derives a cryptographic key from a password using Argon2id.
// Argon2id is a memory-hard key derivation function that provides strong
// resistance against GPU and ASIC attacks.
//
// This is used to derive the Master Encryption Key (MEK) from the user's
// Master Password and Secret Key.
//
// Parameters:
//   - password: User's master password (UTF-8 encoded)
//   - salt: Cryptographic salt (should include User UUID || Secret Key)
//   - params: Argon2 parameters (use DefaultArgon2Params() for recommended values)
//
// Returns:
//   - []byte: Derived key of length params.KeyLength
//   - error: Error if key derivation fails or parameters are invalid
//
// Example:
//
//	params := DefaultArgon2Params()
//	salt := append(userUUID, secretKey...)
//	mek, err := DeriveKey(password, salt, params)
func DeriveKey(password, salt []byte, params *Argon2Params) ([]byte, error) {
	// Validate inputs
	if len(password) == 0 {
		return nil, errors.ErrInvalidPassword
	}

	if len(salt) < int(params.SaltLength) {
		return nil, errors.ErrInvalidSalt
	}

	if params.KeyLength == 0 {
		return nil, errors.ErrInvalidKeyLength
	}

	// Derive key using Argon2id
	// Note: argon2.IDKey uses the "id" variant which is resistant to both
	// side-channel attacks and GPU attacks
	derivedKey := argon2.IDKey(
		password,
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	if len(derivedKey) != int(params.KeyLength) {
		return nil, errors.ErrKeyDerivationFailed
	}

	return derivedKey, nil
}

// DeriveMasterKey is a convenience wrapper for DeriveKey that uses default parameters
// and combines User UUID and Secret Key as the salt.
//
// This implements RootLock's two-secret model:
//   - Master Password (user-chosen, high-entropy)
//   - Secret Key (256-bit random, generated client-side)
//
// Parameters:
//   - password: User's master password
//   - userUUID: User's unique identifier (16 bytes)
//   - secretKey: User's secret key (32 bytes)
//
// Returns:
//   - []byte: Master Encryption Key (MEK) - 32 bytes
//   - error: Error if key derivation fails
//
// Example:
//
//	mek, err := DeriveMasterKey(password, userUUID, secretKey)
//	if err != nil {
//	    return err
//	}
//	// Use MEK to derive item encryption keys via HKDF
func DeriveMasterKey(password, userUUID, secretKey []byte) ([]byte, error) {
	// Validate UUID length (standard UUID is 16 bytes)
	if len(userUUID) != 16 {
		return nil, errors.ErrInvalidSalt
	}

	// Validate Secret Key length (256 bits = 32 bytes)
	if len(secretKey) != 32 {
		return nil, errors.ErrInvalidKeyLength
	}

	// Combine User UUID and Secret Key as salt
	// This ensures that the derived key is unique to both the user and their secret key
	salt := append(userUUID, secretKey...)

	// Use default Argon2id parameters
	params := DefaultArgon2Params()

	return DeriveKey(password, salt, params)
}
