use base64::{engine::general_purpose::STANDARD as BASE64, Engine};
use serde::{Deserialize, Serialize};

use crate::crypto::{
    aes_gcm, argon2, hkdf, random,
    CryptoError,
};

// Re-export API commands
pub use crate::api::*;

#[derive(Debug, Serialize, Deserialize)]
pub struct ApiError {
    pub message: String,
}

impl From<CryptoError> for ApiError {
    fn from(err: CryptoError) -> Self {
        ApiError {
            message: err.to_string(),
        }
    }
}

impl From<base64::DecodeError> for ApiError {
    fn from(err: base64::DecodeError) -> Self {
        ApiError {
            message: format!("Base64 decode error: {}", err),
        }
    }
}

/// Generates a new 256-bit secret key
#[tauri::command]
pub fn generate_secret_key() -> Result<String, ApiError> {
    let secret_key = random::generate_secret_key()?;
    Ok(BASE64.encode(&secret_key))
}

/// Derives the Master Encryption Key from master password, user ID, and secret key
#[tauri::command]
pub fn derive_master_key(
    master_password: String,
    user_id: String,
    secret_key: String, // Base64 encoded
) -> Result<String, ApiError> {
    let secret_key_bytes = BASE64.decode(secret_key)?;
    let mek = argon2::derive_master_key(&master_password, &user_id, &secret_key_bytes)?;
    Ok(BASE64.encode(&mek))
}

/// Derives the Account Unlock Key from the Master Encryption Key
#[tauri::command]
pub fn derive_account_unlock_key(
    master_encryption_key: String, // Base64 encoded
) -> Result<String, ApiError> {
    let mek = BASE64.decode(master_encryption_key)?;
    let auk = hkdf::derive_account_unlock_key(&mek)?;
    Ok(BASE64.encode(&auk))
}

/// Derives the Vault Encryption Key from the Master Encryption Key
#[tauri::command]
pub fn derive_vault_encryption_key(
    master_encryption_key: String, // Base64 encoded
) -> Result<String, ApiError> {
    let mek = BASE64.decode(master_encryption_key)?;
    let vek = hkdf::derive_vault_encryption_key(&mek)?;
    Ok(BASE64.encode(&vek))
}

/// Encrypts data using AES-256-GCM
#[tauri::command]
pub fn encrypt_data(
    plaintext: String,
    key: String, // Base64 encoded
    associated_data: Option<String>,
) -> Result<String, ApiError> {
    let key_bytes = BASE64.decode(key)?;
    let aad = associated_data.as_ref().map(|s| s.as_bytes());
    
    let ciphertext = aes_gcm::encrypt(plaintext.as_bytes(), &key_bytes, aad)?;
    Ok(BASE64.encode(&ciphertext))
}

/// Decrypts data using AES-256-GCM
#[tauri::command]
pub fn decrypt_data(
    ciphertext: String, // Base64 encoded
    key: String, // Base64 encoded
    associated_data: Option<String>,
) -> Result<String, ApiError> {
    let ciphertext_bytes = BASE64.decode(ciphertext)?;
    let key_bytes = BASE64.decode(key)?;
    let aad = associated_data.as_ref().map(|s| s.as_bytes());
    
    let plaintext = aes_gcm::decrypt(&ciphertext_bytes, &key_bytes, aad)?;
    String::from_utf8(plaintext)
        .map_err(|e| ApiError {
            message: format!("Invalid UTF-8: {}", e),
        })
}
