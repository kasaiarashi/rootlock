use argon2::{
    password_hash::{PasswordHasher, SaltString},
    Algorithm, Argon2, Params, Version,
};
use base64::{engine::general_purpose::STANDARD as BASE64, Engine};
use sha2::{Digest, Sha256};

use super::{CryptoError, CryptoResult};

// Argon2id parameters matching backend
const MEMORY_SIZE: u32 = 65536; // 64 MiB
const ITERATIONS: u32 = 3;
const PARALLELISM: u32 = 4;
const OUTPUT_LENGTH: usize = 32; // 32 bytes = 256 bits

/// Derives the Master Encryption Key (MEK) using Argon2id
/// 
/// # Arguments
/// * `master_password` - User's master password
/// * `user_id` - User's UUID (used as part of salt)
/// * `secret_key` - 32-byte secret key (used as part of salt)
/// 
/// # Returns
/// 32-byte Master Encryption Key
pub fn derive_master_key(
    master_password: &str,
    user_id: &str,
    secret_key: &[u8],
) -> CryptoResult<Vec<u8>> {
    if master_password.is_empty() {
        return Err(CryptoError::InvalidInput("Master password cannot be empty".to_string()));
    }
    
    if secret_key.len() != 32 {
        return Err(CryptoError::InvalidInput("Secret key must be 32 bytes".to_string()));
    }

    // Create deterministic salt from user_id + secret_key
    // This ensures the same MEK is derived on different devices
    let mut hasher = Sha256::new();
    hasher.update(user_id.as_bytes());
    hasher.update(secret_key);
    let salt_bytes = hasher.finalize();
    
    // Convert to base64 for SaltString (Argon2 requirement)
    let salt_b64 = BASE64.encode(&salt_bytes[..16]);
    let salt = SaltString::encode_b64(salt_b64.as_bytes())
        .map_err(|e| CryptoError::KeyDerivationError(format!("Invalid salt: {}", e)))?;

    // Configure Argon2id with our parameters
    let params = Params::new(MEMORY_SIZE, ITERATIONS, PARALLELISM, Some(OUTPUT_LENGTH))
        .map_err(|e| CryptoError::KeyDerivationError(format!("Invalid params: {}", e)))?;

    let argon2 = Argon2::new(Algorithm::Argon2id, Version::V0x13, params);

    // Derive the key
    let password_hash = argon2
        .hash_password(master_password.as_bytes(), &salt)
        .map_err(|e| CryptoError::KeyDerivationError(format!("Argon2 failed: {}", e)))?;

    // Extract the hash bytes
    let hash_bytes = password_hash
        .hash
        .ok_or_else(|| CryptoError::KeyDerivationError("No hash generated".to_string()))?;

    Ok(hash_bytes.as_bytes().to_vec())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_derive_master_key() {
        let password = "MySecurePassword123!";
        let user_id = "550e8400-e29b-41d4-a716-446655440000";
        let secret_key = vec![0u8; 32]; // Dummy secret key

        let mek = derive_master_key(password, user_id, &secret_key).unwrap();
        
        assert_eq!(mek.len(), 32);
        
        // Deriving again with same inputs should produce same key
        let mek2 = derive_master_key(password, user_id, &secret_key).unwrap();
        assert_eq!(mek, mek2);
        
        // Different password should produce different key
        let mek3 = derive_master_key("DifferentPassword", user_id, &secret_key).unwrap();
        assert_ne!(mek, mek3);
    }

    #[test]
    fn test_empty_password() {
        let user_id = "550e8400-e29b-41d4-a716-446655440000";
        let secret_key = vec![0u8; 32];

        let result = derive_master_key("", user_id, &secret_key);
        assert!(result.is_err());
    }

    #[test]
    fn test_invalid_secret_key_length() {
        let password = "MySecurePassword123!";
        let user_id = "550e8400-e29b-41d4-a716-446655440000";
        let secret_key = vec![0u8; 16]; // Wrong length

        let result = derive_master_key(password, user_id, &secret_key);
        assert!(result.is_err());
    }
}
