package crypto

import (
	"crypto/sha256"
	"io"

	"golang.org/x/crypto/hkdf"

	"github.com/rootlock/rootlock/backend/pkg/errors"
)

// DeriveKeyHKDF derives a key from input key material using HKDF-SHA256.
// HKDF (HMAC-based Key Derivation Function) is used to expand a master key
// into multiple cryptographically independent keys.
//
// In RootLock, this is used to derive:
//   - Account Unlock Key (AUK) from Master Encryption Key (MEK)
//   - Item Encryption Keys (IEK) from MEK with unique salts per item
//
// Parameters:
//   - ikm: Input key material (e.g., Master Encryption Key)
//   - salt: Optional salt value (can be nil for no salt)
//   - info: Context and application-specific information (e.g., "account-unlock-key")
//   - keyLength: Length of the output key in bytes
//
// Returns:
//   - []byte: Derived key of length keyLength
//   - error: Error if key derivation fails or parameters are invalid
//
// Example:
//
//	// Derive Account Unlock Key from MEK
//	auk, err := DeriveKeyHKDF(mek, nil, []byte("account-unlock-key"), 32)
//
//	// Derive Item Encryption Key with item UUID as salt
//	iek, err := DeriveKeyHKDF(mek, itemUUID, []byte("vault-item-encryption"), 32)
func DeriveKeyHKDF(ikm, salt, info []byte, keyLength int) ([]byte, error) {
	// Validate input key material
	if len(ikm) == 0 {
		return nil, errors.ErrInvalidKeyLength
	}

	// Validate key length
	if keyLength <= 0 || keyLength > 255*sha256.Size {
		// HKDF-SHA256 can generate at most 255 * hash_length bytes
		return nil, errors.ErrInvalidKeyLength
	}

	// Create HKDF reader with SHA-256
	kdf := hkdf.New(sha256.New, ikm, salt, info)

	// Allocate buffer for derived key
	derivedKey := make([]byte, keyLength)

	// Read derived key from HKDF
	if _, err := io.ReadFull(kdf, derivedKey); err != nil {
		return nil, errors.ErrKeyDerivationFailed
	}

	return derivedKey, nil
}

// DeriveAccountUnlockKey derives the Account Unlock Key (AUK) from the Master Encryption Key.
// The AUK is sent to the server (after bcrypt hashing) for authentication.
// It cannot be used to decrypt vault data.
//
// Parameters:
//   - mek: Master Encryption Key (32 bytes)
//
// Returns:
//   - []byte: Account Unlock Key (32 bytes)
//   - error: Error if key derivation fails
//
// Example:
//
//	auk, err := DeriveAccountUnlockKey(mek)
//	aukHash := bcrypt.GenerateFromPassword(auk, bcrypt.DefaultCost)
//	// Send aukHash to server
func DeriveAccountUnlockKey(mek []byte) ([]byte, error) {
	if len(mek) != 32 {
		return nil, errors.ErrInvalidKeyLength
	}

	return DeriveKeyHKDF(
		mek,
		nil, // No salt for AUK
		[]byte("account-unlock-key"),
		32, // 256-bit key
	)
}

// DeriveItemEncryptionKey derives an Item Encryption Key (IEK) from the Master Encryption Key.
// Each vault item has a unique IEK derived using the item's UUID as salt.
// This ensures that compromising one IEK doesn't compromise other items.
//
// Parameters:
//   - mek: Master Encryption Key (32 bytes)
//   - itemUUID: Unique identifier for the vault item (16 bytes)
//
// Returns:
//   - []byte: Item Encryption Key (32 bytes)
//   - error: Error if key derivation fails
//
// Example:
//
//	itemUUID := uuid.New() // Generate new UUID for item
//	iek, err := DeriveItemEncryptionKey(mek, itemUUID[:])
//	encryptedData, err := Encrypt(itemData, iek, itemUUID[:])
func DeriveItemEncryptionKey(mek, itemUUID []byte) ([]byte, error) {
	if len(mek) != 32 {
		return nil, errors.ErrInvalidKeyLength
	}

	if len(itemUUID) != 16 {
		return nil, errors.ErrInvalidSalt
	}

	return DeriveKeyHKDF(
		mek,
		itemUUID, // Use item UUID as salt for uniqueness
		[]byte("vault-item-encryption"),
		32, // 256-bit key
	)
}

// DeriveBackupKey derives a Backup Encryption Key from the Master Encryption Key.
// This can be used for encrypted backups that are stored separately from the main vault.
//
// Parameters:
//   - mek: Master Encryption Key (32 bytes)
//
// Returns:
//   - []byte: Backup Encryption Key (32 bytes)
//   - error: Error if key derivation fails
func DeriveBackupKey(mek []byte) ([]byte, error) {
	if len(mek) != 32 {
		return nil, errors.ErrInvalidKeyLength
	}

	return DeriveKeyHKDF(
		mek,
		nil,
		[]byte("backup-encryption-key"),
		32,
	)
}

// DeriveDeviceKey derives a Device-Specific Key from the Master Encryption Key.
// This can be used for device-specific encryption or authentication.
//
// Parameters:
//   - mek: Master Encryption Key (32 bytes)
//   - deviceID: Unique identifier for the device (16 bytes)
//
// Returns:
//   - []byte: Device Encryption Key (32 bytes)
//   - error: Error if key derivation fails
func DeriveDeviceKey(mek, deviceID []byte) ([]byte, error) {
	if len(mek) != 32 {
		return nil, errors.ErrInvalidKeyLength
	}

	if len(deviceID) != 16 {
		return nil, errors.ErrInvalidSalt
	}

	return DeriveKeyHKDF(
		mek,
		deviceID,
		[]byte("device-encryption-key"),
		32,
	)
}

// DeriveMultipleKeys derives multiple independent keys from a single master key.
// This is useful when you need multiple keys for different purposes.
//
// Parameters:
//   - ikm: Input key material
//   - contexts: Array of context strings (e.g., ["encryption", "authentication", "backup"])
//   - keyLength: Length of each output key in bytes
//
// Returns:
//   - [][]byte: Array of derived keys, one for each context
//   - error: Error if key derivation fails
//
// Example:
//
//	contexts := []string{"encryption-key", "mac-key", "signing-key"}
//	keys, err := DeriveMultipleKeys(mek, contexts, 32)
//	encKey := keys[0]
//	macKey := keys[1]
//	signKey := keys[2]
func DeriveMultipleKeys(ikm []byte, contexts []string, keyLength int) ([][]byte, error) {
	if len(ikm) == 0 {
		return nil, errors.ErrInvalidKeyLength
	}

	if len(contexts) == 0 {
		return nil, errors.ErrKeyDerivationFailed
	}

	keys := make([][]byte, len(contexts))

	for i, context := range contexts {
		key, err := DeriveKeyHKDF(ikm, nil, []byte(context), keyLength)
		if err != nil {
			return nil, err
		}
		keys[i] = key
	}

	return keys, nil
}
