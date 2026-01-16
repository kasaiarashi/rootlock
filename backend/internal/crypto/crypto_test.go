package crypto

import (
	"bytes"
	"testing"
)

// TestGenerateRandomBytes tests random byte generation
func TestGenerateRandomBytes(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		wantErr bool
	}{
		{"Valid 16 bytes", 16, false},
		{"Valid 32 bytes", 32, false},
		{"Valid 64 bytes", 64, false},
		{"Invalid zero length", 0, true},
		{"Invalid negative length", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateRandomBytes(tt.length)

			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateRandomBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(got) != tt.length {
					t.Errorf("GenerateRandomBytes() length = %v, want %v", len(got), tt.length)
				}

				// Test that consecutive calls produce different values (with very high probability)
				got2, _ := GenerateRandomBytes(tt.length)
				if bytes.Equal(got, got2) && tt.length > 0 {
					t.Error("GenerateRandomBytes() produced identical values (extremely unlikely)")
				}
			}
		})
	}
}

// TestGenerateNonce tests nonce generation
func TestGenerateNonce(t *testing.T) {
	nonce, err := GenerateNonce()
	if err != nil {
		t.Fatalf("GenerateNonce() error = %v", err)
	}

	if len(nonce) != NonceSize {
		t.Errorf("GenerateNonce() length = %v, want %v", len(nonce), NonceSize)
	}

	// Test uniqueness
	nonce2, _ := GenerateNonce()
	if bytes.Equal(nonce, nonce2) {
		t.Error("GenerateNonce() produced identical nonces")
	}
}

// TestGenerateSecretKey tests secret key generation
func TestGenerateSecretKey(t *testing.T) {
	key, err := GenerateSecretKey()
	if err != nil {
		t.Fatalf("GenerateSecretKey() error = %v", err)
	}

	if len(key) != 32 {
		t.Errorf("GenerateSecretKey() length = %v, want 32", len(key))
	}

	// Test uniqueness
	key2, _ := GenerateSecretKey()
	if bytes.Equal(key, key2) {
		t.Error("GenerateSecretKey() produced identical keys")
	}
}

// TestArgon2KeyDerivation tests Argon2id key derivation
func TestArgon2KeyDerivation(t *testing.T) {
	password := []byte("TestPassword123!")
	salt := []byte("1234567890123456") // 16 bytes
	params := DefaultArgon2Params()

	key, err := DeriveKey(password, salt, params)
	if err != nil {
		t.Fatalf("DeriveKey() error = %v", err)
	}

	if len(key) != int(params.KeyLength) {
		t.Errorf("DeriveKey() length = %v, want %v", len(key), params.KeyLength)
	}

	// Test determinism (same inputs should produce same output)
	key2, err := DeriveKey(password, salt, params)
	if err != nil {
		t.Fatalf("DeriveKey() second call error = %v", err)
	}

	if !bytes.Equal(key, key2) {
		t.Error("DeriveKey() not deterministic")
	}

	// Test that different passwords produce different keys
	differentPassword := []byte("DifferentPassword")
	key3, err := DeriveKey(differentPassword, salt, params)
	if err != nil {
		t.Fatalf("DeriveKey() with different password error = %v", err)
	}

	if bytes.Equal(key, key3) {
		t.Error("DeriveKey() produced same key for different passwords")
	}
}

// TestDeriveMasterKey tests the convenience wrapper for master key derivation
func TestDeriveMasterKey(t *testing.T) {
	password := []byte("MyMasterPassword123!")
	userUUID := []byte("1234567890123456")                  // 16 bytes
	secretKey := []byte("12345678901234567890123456789012") // 32 bytes

	mek, err := DeriveMasterKey(password, userUUID, secretKey)
	if err != nil {
		t.Fatalf("DeriveMasterKey() error = %v", err)
	}

	if len(mek) != 32 {
		t.Errorf("DeriveMasterKey() length = %v, want 32", len(mek))
	}

	// Test with invalid UUID length
	invalidUUID := []byte("short")
	_, err = DeriveMasterKey(password, invalidUUID, secretKey)
	if err == nil {
		t.Error("DeriveMasterKey() should fail with invalid UUID length")
	}

	// Test with invalid secret key length
	invalidKey := []byte("short")
	_, err = DeriveMasterKey(password, userUUID, invalidKey)
	if err == nil {
		t.Error("DeriveMasterKey() should fail with invalid secret key length")
	}
}

// TestAESGCMEncryption tests AES-256-GCM encryption and decryption
func TestAESGCMEncryption(t *testing.T) {
	key := make([]byte, AESKeySize)
	for i := range key {
		key[i] = byte(i)
	}

	plaintext := []byte("Hello, RootLock! This is a secret message.")
	aad := []byte("additional-authenticated-data")

	// Test encryption
	ciphertext, err := Encrypt(plaintext, key, aad)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Ciphertext should be longer than plaintext (nonce + tag)
	expectedLength := len(plaintext) + NonceSize + TagSize
	if len(ciphertext) != expectedLength {
		t.Errorf("Encrypt() ciphertext length = %v, want %v", len(ciphertext), expectedLength)
	}

	// Test decryption
	decrypted, err := Decrypt(ciphertext, key, aad)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Errorf("Decrypt() = %v, want %v", string(decrypted), string(plaintext))
	}

	// Test that decryption fails with wrong key
	wrongKey := make([]byte, AESKeySize)
	_, err = Decrypt(ciphertext, wrongKey, aad)
	if err == nil {
		t.Error("Decrypt() should fail with wrong key")
	}

	// Test that decryption fails with wrong AAD
	wrongAAD := []byte("wrong-aad")
	_, err = Decrypt(ciphertext, key, wrongAAD)
	if err == nil {
		t.Error("Decrypt() should fail with wrong AAD")
	}

	// Test that decryption fails with tampered ciphertext
	tamperedCiphertext := make([]byte, len(ciphertext))
	copy(tamperedCiphertext, ciphertext)
	tamperedCiphertext[len(tamperedCiphertext)-1] ^= 0xFF // Flip bits in tag
	_, err = Decrypt(tamperedCiphertext, key, aad)
	if err == nil {
		t.Error("Decrypt() should fail with tampered ciphertext")
	}
}

// TestAESGCMWithEmptyPlaintext tests encryption of empty data
func TestAESGCMWithEmptyPlaintext(t *testing.T) {
	key := make([]byte, AESKeySize)
	plaintext := []byte{}
	aad := []byte("metadata")

	ciphertext, err := Encrypt(plaintext, key, aad)
	if err != nil {
		t.Fatalf("Encrypt() empty plaintext error = %v", err)
	}

	decrypted, err := Decrypt(ciphertext, key, aad)
	if err != nil {
		t.Fatalf("Decrypt() empty plaintext error = %v", err)
	}

	if !bytes.Equal(plaintext, decrypted) {
		t.Error("Decrypt() failed for empty plaintext")
	}
}

// TestHKDFDerivation tests HKDF key derivation
func TestHKDFDerivation(t *testing.T) {
	ikm := make([]byte, 32) // Input key material
	for i := range ikm {
		ikm[i] = byte(i)
	}

	salt := []byte("random-salt")
	info := []byte("application-context")

	key, err := DeriveKeyHKDF(ikm, salt, info, 32)
	if err != nil {
		t.Fatalf("DeriveKeyHKDF() error = %v", err)
	}

	if len(key) != 32 {
		t.Errorf("DeriveKeyHKDF() length = %v, want 32", len(key))
	}

	// Test determinism
	key2, err := DeriveKeyHKDF(ikm, salt, info, 32)
	if err != nil {
		t.Fatalf("DeriveKeyHKDF() second call error = %v", err)
	}

	if !bytes.Equal(key, key2) {
		t.Error("DeriveKeyHKDF() not deterministic")
	}

	// Test that different info produces different keys
	differentInfo := []byte("different-context")
	key3, err := DeriveKeyHKDF(ikm, salt, differentInfo, 32)
	if err != nil {
		t.Fatalf("DeriveKeyHKDF() with different info error = %v", err)
	}

	if bytes.Equal(key, key3) {
		t.Error("DeriveKeyHKDF() produced same key for different info")
	}

	// Test nil salt (should work)
	key4, err := DeriveKeyHKDF(ikm, nil, info, 32)
	if err != nil {
		t.Fatalf("DeriveKeyHKDF() with nil salt error = %v", err)
	}

	if len(key4) != 32 {
		t.Error("DeriveKeyHKDF() failed with nil salt")
	}
}

// TestDeriveAccountUnlockKey tests AUK derivation
func TestDeriveAccountUnlockKey(t *testing.T) {
	mek := make([]byte, 32)
	for i := range mek {
		mek[i] = byte(i)
	}

	auk, err := DeriveAccountUnlockKey(mek)
	if err != nil {
		t.Fatalf("DeriveAccountUnlockKey() error = %v", err)
	}

	if len(auk) != 32 {
		t.Errorf("DeriveAccountUnlockKey() length = %v, want 32", len(auk))
	}

	// Test that AUK is different from MEK
	if bytes.Equal(mek, auk) {
		t.Error("DeriveAccountUnlockKey() should produce different key from MEK")
	}

	// Test invalid MEK length
	invalidMEK := []byte("short")
	_, err = DeriveAccountUnlockKey(invalidMEK)
	if err == nil {
		t.Error("DeriveAccountUnlockKey() should fail with invalid MEK length")
	}
}

// TestDeriveItemEncryptionKey tests IEK derivation
func TestDeriveItemEncryptionKey(t *testing.T) {
	mek := make([]byte, 32)
	itemUUID := make([]byte, 16)

	iek, err := DeriveItemEncryptionKey(mek, itemUUID)
	if err != nil {
		t.Fatalf("DeriveItemEncryptionKey() error = %v", err)
	}

	if len(iek) != 32 {
		t.Errorf("DeriveItemEncryptionKey() length = %v, want 32", len(iek))
	}

	// Test that different UUIDs produce different keys
	itemUUID2 := make([]byte, 16)
	itemUUID2[0] = 1 // Make it different

	iek2, err := DeriveItemEncryptionKey(mek, itemUUID2)
	if err != nil {
		t.Fatalf("DeriveItemEncryptionKey() with different UUID error = %v", err)
	}

	if bytes.Equal(iek, iek2) {
		t.Error("DeriveItemEncryptionKey() produced same key for different UUIDs")
	}
}

// TestEndToEndWorkflow tests the complete encryption workflow
func TestEndToEndWorkflow(t *testing.T) {
	// Step 1: User provides master password and has UUID + Secret Key
	masterPassword := []byte("MySecureMasterPassword123!")
	userUUID := []byte("1234567890123456") // 16 bytes
	secretKey := make([]byte, 32)          // 32 bytes (generated during registration)
	copy(secretKey, "12345678901234567890123456789012")

	// Step 2: Derive Master Encryption Key (MEK)
	mek, err := DeriveMasterKey(masterPassword, userUUID, secretKey)
	if err != nil {
		t.Fatalf("DeriveMasterKey() error = %v", err)
	}

	// Step 3: Derive Account Unlock Key (AUK) for server authentication
	auk, err := DeriveAccountUnlockKey(mek)
	if err != nil {
		t.Fatalf("DeriveAccountUnlockKey() error = %v", err)
	}

	// Step 4: User creates a vault item
	itemUUID := []byte("abcdef1234567890") // 16 bytes
	itemData := []byte(`{"username":"user@example.com","password":"secret123"}`)

	// Step 5: Derive Item Encryption Key (IEK)
	iek, err := DeriveItemEncryptionKey(mek, itemUUID)
	if err != nil {
		t.Fatalf("DeriveItemEncryptionKey() error = %v", err)
	}

	// Step 6: Encrypt item with IEK (using item UUID as AAD)
	encryptedItem, err := Encrypt(itemData, iek, itemUUID)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Step 7: Later, decrypt item (simulating retrieval from server)
	// Re-derive MEK from password
	mek2, err := DeriveMasterKey(masterPassword, userUUID, secretKey)
	if err != nil {
		t.Fatalf("Re-derivation of MEK error = %v", err)
	}

	// Re-derive IEK
	iek2, err := DeriveItemEncryptionKey(mek2, itemUUID)
	if err != nil {
		t.Fatalf("Re-derivation of IEK error = %v", err)
	}

	// Decrypt item
	decryptedItem, err := Decrypt(encryptedItem, iek2, itemUUID)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	// Verify decrypted data matches original
	if !bytes.Equal(itemData, decryptedItem) {
		t.Errorf("Decrypted data = %v, want %v", string(decryptedItem), string(itemData))
	}

	// Verify AUK is consistent
	auk2, err := DeriveAccountUnlockKey(mek2)
	if err != nil {
		t.Fatalf("Re-derivation of AUK error = %v", err)
	}

	if !bytes.Equal(auk, auk2) {
		t.Error("Re-derived AUK doesn't match original")
	}

	t.Log("✓ End-to-end workflow successful")
}

// BenchmarkArgon2 benchmarks Argon2id key derivation
func BenchmarkArgon2(b *testing.B) {
	password := []byte("BenchmarkPassword")
	salt := make([]byte, 16)
	params := DefaultArgon2Params()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DeriveKey(password, salt, params)
	}
}

// BenchmarkAESGCMEncrypt benchmarks AES-GCM encryption
func BenchmarkAESGCMEncrypt(b *testing.B) {
	key := make([]byte, AESKeySize)
	plaintext := make([]byte, 1024) // 1KB
	aad := []byte("benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Encrypt(plaintext, key, aad)
	}
}

// BenchmarkAESGCMDecrypt benchmarks AES-GCM decryption
func BenchmarkAESGCMDecrypt(b *testing.B) {
	key := make([]byte, AESKeySize)
	plaintext := make([]byte, 1024) // 1KB
	aad := []byte("benchmark")

	ciphertext, _ := Encrypt(plaintext, key, aad)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Decrypt(ciphertext, key, aad)
	}
}

// BenchmarkHKDF benchmarks HKDF key derivation
func BenchmarkHKDF(b *testing.B) {
	ikm := make([]byte, 32)
	info := []byte("benchmark-context")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DeriveKeyHKDF(ikm, nil, info, 32)
	}
}
