package errors

import "errors"

// Cryptographic errors
var (
	ErrInvalidKeyLength     = errors.New("invalid key length")
	ErrInvalidNonceLength   = errors.New("invalid nonce length")
	ErrEncryptionFailed     = errors.New("encryption failed")
	ErrDecryptionFailed     = errors.New("decryption failed")
	ErrAuthenticationFailed = errors.New("authentication failed")
	ErrInvalidCiphertext    = errors.New("invalid ciphertext")
	ErrInvalidSaltLength    = errors.New("invalid salt length")
)

// Key derivation errors
var (
	ErrKeyDerivationFailed = errors.New("key derivation failed")
	ErrInvalidPassword     = errors.New("invalid password")
	ErrInvalidSalt         = errors.New("invalid salt")
)

// Random generation errors
var (
	ErrRandomGenerationFailed = errors.New("random generation failed")
	ErrInsufficientEntropy    = errors.New("insufficient entropy")
)
