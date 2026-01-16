pub mod argon2;
pub mod aes_gcm;
pub mod hkdf;
pub mod random;

pub use self::argon2::*;
pub use self::aes_gcm::*;
pub use self::hkdf::*;
pub use self::random::*;

use serde::{Deserialize, Serialize};

#[derive(Debug, thiserror::Error)]
pub enum CryptoError {
    #[error("Encryption failed: {0}")]
    EncryptionError(String),
    
    #[error("Decryption failed: {0}")]
    DecryptionError(String),
    
    #[error("Key derivation failed: {0}")]
    KeyDerivationError(String),
    
    #[error("Invalid input: {0}")]
    InvalidInput(String),
    
    #[error("Base64 error: {0}")]
    Base64Error(#[from] base64::DecodeError),
}

pub type CryptoResult<T> = Result<T, CryptoError>;

// Request/Response types for Tauri commands
#[derive(Debug, Serialize, Deserialize)]
pub struct DeriveMasterKeyRequest {
    pub master_password: String,
    pub user_id: String,
    pub secret_key: String, // Base64 encoded
}

#[derive(Debug, Serialize, Deserialize)]
pub struct DeriveMasterKeyResponse {
    pub master_encryption_key: String, // Base64 encoded
}

#[derive(Debug, Serialize, Deserialize)]
pub struct DeriveAukRequest {
    pub master_encryption_key: String, // Base64 encoded
}

#[derive(Debug, Serialize, Deserialize)]
pub struct DeriveAukResponse {
    pub account_unlock_key: String, // Base64 encoded
}

#[derive(Debug, Serialize, Deserialize)]
pub struct EncryptRequest {
    pub plaintext: String,
    pub key: String, // Base64 encoded
    pub associated_data: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct EncryptResponse {
    pub ciphertext: String, // Base64 encoded (nonce + encrypted data + tag)
}

#[derive(Debug, Serialize, Deserialize)]
pub struct DecryptRequest {
    pub ciphertext: String, // Base64 encoded
    pub key: String, // Base64 encoded
    pub associated_data: Option<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct DecryptResponse {
    pub plaintext: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct GenerateSecretKeyResponse {
    pub secret_key: String, // Base64 encoded 32-byte key
}
