use hkdf::Hkdf;
use sha2::Sha256;

use super::{CryptoError, CryptoResult};

const OUTPUT_KEY_LENGTH: usize = 32; // 256 bits

/// Derives the Account Unlock Key (AUK) from the Master Encryption Key
/// 
/// # Arguments
/// * `mek` - Master Encryption Key (32 bytes)
/// 
/// # Returns
/// 32-byte Account Unlock Key
pub fn derive_account_unlock_key(mek: &[u8]) -> CryptoResult<Vec<u8>> {
    if mek.len() != 32 {
        return Err(CryptoError::InvalidInput("MEK must be 32 bytes".to_string()));
    }

    let hkdf = Hkdf::<Sha256>::new(None, mek);
    let mut okm = vec![0u8; OUTPUT_KEY_LENGTH];
    
    hkdf.expand(b"account-unlock-key", &mut okm)
        .map_err(|e| CryptoError::KeyDerivationError(format!("HKDF expand failed: {}", e)))?;

    Ok(okm)
}

/// Derives an Item Encryption Key (IEK) from the Master Encryption Key
/// 
/// # Arguments
/// * `mek` - Master Encryption Key (32 bytes)
/// * `item_id` - Unique identifier for the vault item
/// 
/// # Returns
/// 32-byte Item Encryption Key
pub fn derive_item_encryption_key(mek: &[u8], item_id: &str) -> CryptoResult<Vec<u8>> {
    if mek.len() != 32 {
        return Err(CryptoError::InvalidInput("MEK must be 32 bytes".to_string()));
    }

    let hkdf = Hkdf::<Sha256>::new(None, mek);
    let mut okm = vec![0u8; OUTPUT_KEY_LENGTH];
    
    // Use item_id as context for key derivation
    let info = format!("item-encryption-key:{}", item_id);
    
    hkdf.expand(info.as_bytes(), &mut okm)
        .map_err(|e| CryptoError::KeyDerivationError(format!("HKDF expand failed: {}", e)))?;

    Ok(okm)
}

/// Derives the Vault Encryption Key (VEK) from the Master Encryption Key
/// Used to encrypt the entire vault blob before sending to server
/// 
/// # Arguments
/// * `mek` - Master Encryption Key (32 bytes)
/// 
/// # Returns
/// 32-byte Vault Encryption Key
pub fn derive_vault_encryption_key(mek: &[u8]) -> CryptoResult<Vec<u8>> {
    if mek.len() != 32 {
        return Err(CryptoError::InvalidInput("MEK must be 32 bytes".to_string()));
    }

    let hkdf = Hkdf::<Sha256>::new(None, mek);
    let mut okm = vec![0u8; OUTPUT_KEY_LENGTH];
    
    hkdf.expand(b"vault-encryption-key", &mut okm)
        .map_err(|e| CryptoError::KeyDerivationError(format!("HKDF expand failed: {}", e)))?;

    Ok(okm)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_derive_account_unlock_key() {
        let mek = vec![0u8; 32];
        
        let auk = derive_account_unlock_key(&mek).unwrap();
        assert_eq!(auk.len(), 32);
        
        // Should be deterministic
        let auk2 = derive_account_unlock_key(&mek).unwrap();
        assert_eq!(auk, auk2);
    }

    #[test]
    fn test_derive_item_encryption_key() {
        let mek = vec![0u8; 32];
        let item_id = "550e8400-e29b-41d4-a716-446655440000";
        
        let iek = derive_item_encryption_key(&mek, item_id).unwrap();
        assert_eq!(iek.len(), 32);
        
        // Should be deterministic for same item_id
        let iek2 = derive_item_encryption_key(&mek, item_id).unwrap();
        assert_eq!(iek, iek2);
        
        // Different item_id should produce different key
        let iek3 = derive_item_encryption_key(&mek, "different-id").unwrap();
        assert_ne!(iek, iek3);
    }

    #[test]
    fn test_derive_vault_encryption_key() {
        let mek = vec![0u8; 32];
        
        let vek = derive_vault_encryption_key(&mek).unwrap();
        assert_eq!(vek.len(), 32);
        
        // Should be deterministic
        let vek2 = derive_vault_encryption_key(&mek).unwrap();
        assert_eq!(vek, vek2);
    }

    #[test]
    fn test_different_derived_keys() {
        let mek = vec![0u8; 32];
        
        let auk = derive_account_unlock_key(&mek).unwrap();
        let vek = derive_vault_encryption_key(&mek).unwrap();
        let iek = derive_item_encryption_key(&mek, "test-item").unwrap();
        
        // All derived keys should be different
        assert_ne!(auk, vek);
        assert_ne!(auk, iek);
        assert_ne!(vek, iek);
    }

    #[test]
    fn test_invalid_mek_length() {
        let mek = vec![0u8; 16]; // Wrong length
        
        assert!(derive_account_unlock_key(&mek).is_err());
        assert!(derive_vault_encryption_key(&mek).is_err());
        assert!(derive_item_encryption_key(&mek, "test").is_err());
    }

    #[test]
    fn test_different_meks_produce_different_keys() {
        let mek1 = vec![1u8; 32];
        let mek2 = vec![2u8; 32];
        
        let auk1 = derive_account_unlock_key(&mek1).unwrap();
        let auk2 = derive_account_unlock_key(&mek2).unwrap();
        
        assert_ne!(auk1, auk2);
    }
}
