use aes_gcm::{
    aead::{Aead, KeyInit, Payload},
    Aes256Gcm, Nonce,
};

use super::{random::generate_random_bytes, CryptoError, CryptoResult};

const NONCE_SIZE: usize = 12; // 96 bits recommended for GCM
const KEY_SIZE: usize = 32; // 256 bits

/// Encrypts plaintext using AES-256-GCM
/// 
/// # Arguments
/// * `plaintext` - Data to encrypt
/// * `key` - 32-byte encryption key
/// * `associated_data` - Optional additional authenticated data (AAD)
/// 
/// # Returns
/// Encrypted data formatted as: nonce || ciphertext || tag
pub fn encrypt(
    plaintext: &[u8],
    key: &[u8],
    associated_data: Option<&[u8]>,
) -> CryptoResult<Vec<u8>> {
    if key.len() != KEY_SIZE {
        return Err(CryptoError::InvalidInput(format!(
            "Key must be {} bytes, got {}",
            KEY_SIZE,
            key.len()
        )));
    }

    // Generate random nonce
    let nonce_bytes = generate_random_bytes(NONCE_SIZE)?;
    let nonce = Nonce::from_slice(&nonce_bytes);

    // Create cipher
    let cipher = Aes256Gcm::new_from_slice(key)
        .map_err(|e| CryptoError::EncryptionError(format!("Invalid key: {}", e)))?;

    // Prepare payload with optional AAD
    let payload = match associated_data {
        Some(aad) => Payload {
            msg: plaintext,
            aad,
        },
        None => Payload {
            msg: plaintext,
            aad: &[],
        },
    };

    // Encrypt
    let ciphertext = cipher
        .encrypt(nonce, payload)
        .map_err(|e| CryptoError::EncryptionError(format!("Encryption failed: {}", e)))?;

    // Combine nonce + ciphertext (ciphertext already includes auth tag)
    let mut result = Vec::with_capacity(NONCE_SIZE + ciphertext.len());
    result.extend_from_slice(&nonce_bytes);
    result.extend_from_slice(&ciphertext);

    Ok(result)
}

/// Decrypts ciphertext using AES-256-GCM
/// 
/// # Arguments
/// * `ciphertext` - Encrypted data (nonce || ciphertext || tag)
/// * `key` - 32-byte decryption key
/// * `associated_data` - Optional additional authenticated data (AAD) - must match encryption
/// 
/// # Returns
/// Decrypted plaintext
pub fn decrypt(
    ciphertext: &[u8],
    key: &[u8],
    associated_data: Option<&[u8]>,
) -> CryptoResult<Vec<u8>> {
    if key.len() != KEY_SIZE {
        return Err(CryptoError::InvalidInput(format!(
            "Key must be {} bytes, got {}",
            KEY_SIZE,
            key.len()
        )));
    }

    if ciphertext.len() < NONCE_SIZE {
        return Err(CryptoError::DecryptionError(
            "Ciphertext too short".to_string(),
        ));
    }

    // Extract nonce and encrypted data
    let (nonce_bytes, encrypted_data) = ciphertext.split_at(NONCE_SIZE);
    let nonce = Nonce::from_slice(nonce_bytes);

    // Create cipher
    let cipher = Aes256Gcm::new_from_slice(key)
        .map_err(|e| CryptoError::DecryptionError(format!("Invalid key: {}", e)))?;

    // Prepare payload with optional AAD
    let payload = match associated_data {
        Some(aad) => Payload {
            msg: encrypted_data,
            aad,
        },
        None => Payload {
            msg: encrypted_data,
            aad: &[],
        },
    };

    // Decrypt and verify
    let plaintext = cipher
        .decrypt(nonce, payload)
        .map_err(|e| CryptoError::DecryptionError(format!("Decryption failed (wrong key or tampered data): {}", e)))?;

    Ok(plaintext)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_encrypt_decrypt() {
        let key = vec![0u8; 32];
        let plaintext = b"Hello, World!";

        let ciphertext = encrypt(plaintext, &key, None).unwrap();
        assert!(ciphertext.len() > plaintext.len()); // Should be longer due to nonce and tag

        let decrypted = decrypt(&ciphertext, &key, None).unwrap();
        assert_eq!(plaintext, decrypted.as_slice());
    }

    #[test]
    fn test_encrypt_decrypt_with_aad() {
        let key = vec![0u8; 32];
        let plaintext = b"Secret message";
        let aad = b"context-info";

        let ciphertext = encrypt(plaintext, &key, Some(aad)).unwrap();
        let decrypted = decrypt(&ciphertext, &key, Some(aad)).unwrap();
        
        assert_eq!(plaintext, decrypted.as_slice());

        // Decryption with wrong AAD should fail
        let wrong_aad = b"wrong-context";
        let result = decrypt(&ciphertext, &key, Some(wrong_aad));
        assert!(result.is_err());
    }

    #[test]
    fn test_tampered_ciphertext() {
        let key = vec![0u8; 32];
        let plaintext = b"Important data";

        let mut ciphertext = encrypt(plaintext, &key, None).unwrap();
        
        // Tamper with the ciphertext
        ciphertext[20] ^= 0xFF;

        // Decryption should fail due to authentication
        let result = decrypt(&ciphertext, &key, None);
        assert!(result.is_err());
    }

    #[test]
    fn test_wrong_key() {
        let key1 = vec![1u8; 32];
        let key2 = vec![2u8; 32];
        let plaintext = b"Confidential";

        let ciphertext = encrypt(plaintext, &key1, None).unwrap();
        
        // Decryption with wrong key should fail
        let result = decrypt(&ciphertext, &key2, None);
        assert!(result.is_err());
    }

    #[test]
    fn test_invalid_key_length() {
        let key = vec![0u8; 16]; // Wrong length
        let plaintext = b"Test";

        let result = encrypt(plaintext, &key, None);
        assert!(result.is_err());
    }

    #[test]
    fn test_nonce_uniqueness() {
        let key = vec![0u8; 32];
        let plaintext = b"Same plaintext";

        let ciphertext1 = encrypt(plaintext, &key, None).unwrap();
        let ciphertext2 = encrypt(plaintext, &key, None).unwrap();

        // Nonces should be different (first 12 bytes)
        assert_ne!(&ciphertext1[..12], &ciphertext2[..12]);
        
        // But both should decrypt to same plaintext
        assert_eq!(decrypt(&ciphertext1, &key, None).unwrap(), plaintext);
        assert_eq!(decrypt(&ciphertext2, &key, None).unwrap(), plaintext);
    }
}
