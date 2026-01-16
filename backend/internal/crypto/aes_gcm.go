package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/subtle"

	"github.com/rootlock/rootlock/backend/pkg/errors"
)

const (
	// AESKeySize is the key size for AES-256 in bytes
	AESKeySize = 32

	// NonceSize is the nonce size for AES-GCM in bytes (96 bits)
	NonceSize = 12

	// TagSize is the authentication tag size for AES-GCM in bytes (128 bits)
	TagSize = 16
)

// Encrypt encrypts plaintext using AES-256-GCM with the provided key and additional data.
// This provides authenticated encryption, ensuring both confidentiality and integrity.
//
// The ciphertext format is: nonce || ciphertext || tag
// - First 12 bytes: Random nonce
// - Middle bytes: Encrypted data
// - Last 16 bytes: Authentication tag
//
// Parameters:
//   - plaintext: Data to encrypt
//   - key: 256-bit encryption key (32 bytes)
//   - additionalData: Additional authenticated data (AAD) - not encrypted but authenticated
//
// Returns:
//   - []byte: Encrypted data (nonce || ciphertext || tag)
//   - error: Error if encryption fails or key is invalid
//
// Security Notes:
//   - The nonce MUST be unique for each encryption with the same key
//   - The nonce is generated automatically using crypto/rand
//   - Never reuse a nonce with the same key (catastrophic security failure)
//
// Example:
//
//	key := []byte{...} // 32-byte key
//	plaintext := []byte("secret data")
//	aad := []byte("item-uuid-v4")
//	ciphertext, err := Encrypt(plaintext, key, aad)
func Encrypt(plaintext, key, additionalData []byte) ([]byte, error) {
	// Validate key length (must be 256 bits for AES-256)
	if len(key) != AESKeySize {
		return nil, errors.ErrInvalidKeyLength
	}

	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.ErrEncryptionFailed
	}

	// Create GCM mode cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.ErrEncryptionFailed
	}

	// Generate cryptographically secure random nonce
	nonce, err := GenerateNonce()
	if err != nil {
		return nil, err
	}

	// Encrypt and authenticate
	// GCM.Seal appends the ciphertext and tag to the nonce
	ciphertext := gcm.Seal(nonce, nonce, plaintext, additionalData)

	return ciphertext, nil
}

// Decrypt decrypts ciphertext using AES-256-GCM with the provided key and additional data.
// This verifies both the integrity and authenticity of the data before decrypting.
//
// The ciphertext format must be: nonce || ciphertext || tag
//
// Parameters:
//   - ciphertext: Encrypted data (nonce || ciphertext || tag)
//   - key: 256-bit encryption key (32 bytes) - must match encryption key
//   - additionalData: Additional authenticated data (AAD) - must match encryption AAD
//
// Returns:
//   - []byte: Decrypted plaintext
//   - error: Error if decryption fails, authentication fails, or ciphertext is invalid
//
// Security Notes:
//   - If the authentication tag is invalid, an error is returned
//   - If the AAD doesn't match, authentication fails
//   - Timing-safe comparison is used to prevent timing attacks
//
// Example:
//
//	key := []byte{...} // 32-byte key (same as encryption)
//	aad := []byte("item-uuid-v4") // Same as encryption
//	plaintext, err := Decrypt(ciphertext, key, aad)
//	if err != nil {
//	    // Authentication failed or invalid ciphertext
//	    return err
//	}
func Decrypt(ciphertext, key, additionalData []byte) ([]byte, error) {
	// Validate key length
	if len(key) != AESKeySize {
		return nil, errors.ErrInvalidKeyLength
	}

	// Validate ciphertext length (must contain at least nonce + tag)
	if len(ciphertext) < NonceSize+TagSize {
		return nil, errors.ErrInvalidCiphertext
	}

	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.ErrDecryptionFailed
	}

	// Create GCM mode cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.ErrDecryptionFailed
	}

	// Extract nonce from the beginning of ciphertext
	nonce := ciphertext[:NonceSize]
	encryptedData := ciphertext[NonceSize:]

	// Decrypt and verify authentication tag
	// GCM.Open verifies the tag before decrypting (constant-time)
	plaintext, err := gcm.Open(nil, nonce, encryptedData, additionalData)
	if err != nil {
		// Authentication failed or decryption failed
		return nil, errors.ErrAuthenticationFailed
	}

	return plaintext, nil
}

// EncryptWithNonce encrypts plaintext with a provided nonce instead of generating one.
// This should only be used in specific cases where nonce management is handled externally.
//
// WARNING: Using this function incorrectly can lead to nonce reuse, which completely
// breaks AES-GCM security. Only use this if you have a strong reason and understand
// the security implications.
//
// Parameters:
//   - plaintext: Data to encrypt
//   - key: 256-bit encryption key (32 bytes)
//   - nonce: 96-bit nonce (12 bytes) - MUST be unique for this key
//   - additionalData: Additional authenticated data (AAD)
//
// Returns:
//   - []byte: Encrypted data (nonce || ciphertext || tag)
//   - error: Error if encryption fails or parameters are invalid
func EncryptWithNonce(plaintext, key, nonce, additionalData []byte) ([]byte, error) {
	// Validate key length
	if len(key) != AESKeySize {
		return nil, errors.ErrInvalidKeyLength
	}

	// Validate nonce length
	if len(nonce) != NonceSize {
		return nil, errors.ErrInvalidNonceLength
	}

	// Create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.ErrEncryptionFailed
	}

	// Create GCM mode cipher
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.ErrEncryptionFailed
	}

	// Encrypt and authenticate with provided nonce
	ciphertext := gcm.Seal(nonce, nonce, plaintext, additionalData)

	return ciphertext, nil
}

// VerifyTag verifies that the authentication tag in the ciphertext is valid
// without decrypting the data. This can be used to quickly verify integrity
// before attempting decryption.
//
// Parameters:
//   - ciphertext: Encrypted data (nonce || ciphertext || tag)
//   - key: 256-bit encryption key (32 bytes)
//   - additionalData: Additional authenticated data (AAD)
//
// Returns:
//   - bool: true if tag is valid, false otherwise
func VerifyTag(ciphertext, key, additionalData []byte) bool {
	// Attempt decryption (which verifies the tag)
	_, err := Decrypt(ciphertext, key, additionalData)

	// If decryption succeeds, tag is valid
	return err == nil
}

// CompareKeys compares two encryption keys in constant time to prevent timing attacks.
// Use this instead of bytes.Equal() when comparing sensitive keys.
//
// Parameters:
//   - key1: First key
//   - key2: Second key
//
// Returns:
//   - bool: true if keys are equal, false otherwise
func CompareKeys(key1, key2 []byte) bool {
	// subtle.ConstantTimeCompare returns 1 if equal, 0 otherwise
	return subtle.ConstantTimeCompare(key1, key2) == 1
}
