use rand::RngCore;

use super::{CryptoError, CryptoResult};

/// Generates cryptographically secure random bytes
/// 
/// # Arguments
/// * `length` - Number of random bytes to generate
/// 
/// # Returns
/// Vector of random bytes
pub fn generate_random_bytes(length: usize) -> CryptoResult<Vec<u8>> {
    if length == 0 {
        return Err(CryptoError::InvalidInput("Length must be greater than 0".to_string()));
    }

    let mut bytes = vec![0u8; length];
    rand::thread_rng()
        .try_fill_bytes(&mut bytes)
        .map_err(|e| CryptoError::InvalidInput(format!("Random generation failed: {}", e)))?;

    Ok(bytes)
}

/// Generates a 256-bit (32-byte) secret key
/// This is used as the user's Secret Key in the two-secret architecture
/// 
/// # Returns
/// 32-byte random secret key
pub fn generate_secret_key() -> CryptoResult<Vec<u8>> {
    generate_random_bytes(32)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_generate_random_bytes() {
        let bytes = generate_random_bytes(32).unwrap();
        assert_eq!(bytes.len(), 32);
        
        // Generate again and ensure it's different
        let bytes2 = generate_random_bytes(32).unwrap();
        assert_ne!(bytes, bytes2);
    }

    #[test]
    fn test_generate_secret_key() {
        let key = generate_secret_key().unwrap();
        assert_eq!(key.len(), 32);
        
        // Should be different each time
        let key2 = generate_secret_key().unwrap();
        assert_ne!(key, key2);
    }

    #[test]
    fn test_zero_length() {
        let result = generate_random_bytes(0);
        assert!(result.is_err());
    }

    #[test]
    fn test_various_lengths() {
        for length in [1, 16, 32, 64, 128].iter() {
            let bytes = generate_random_bytes(*length).unwrap();
            assert_eq!(bytes.len(), *length);
        }
    }
}
