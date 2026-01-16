use reqwest::blocking::Client;
use serde::{Deserialize, Serialize};
use std::time::Duration;

use crate::commands::ApiError;

const API_BASE_URL: &str = "http://localhost:8080/api/v1";
const TIMEOUT_SECONDS: u64 = 30;

// Request/Response types
#[derive(Debug, Serialize, Deserialize)]
pub struct RegisterRequest {
    pub email: String,
    pub auk_hash: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct RegisterResponse {
    pub message: String,
    pub user: UserInfo,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct LoginRequest {
    pub email: String,
    pub auk: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct LoginResponse {
    pub user: UserInfo,
    pub tokens: TokenPair,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct UserInfo {
    pub id: String,
    pub email: String,
    pub created_at: String,
    pub last_login: Option<String>,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct TokenPair {
    pub access_token: String,
    pub refresh_token: String,
    pub expires_in: i64,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct VaultResponse {
    pub id: String,
    pub encrypted_blob: String, // Base64 encoded
    pub version: i32,
    pub last_modified: String,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct UpdateVaultRequest {
    pub encrypted_blob: String, // Base64 encoded
    pub version: i32,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct UpdateVaultResponse {
    pub id: String,
    pub version: i32,
    pub last_modified: String,
    pub message: String,
}

/// Register a new user
#[tauri::command]
pub fn register_user(email: String, auk_hash: String) -> Result<RegisterResponse, ApiError> {
    let client = Client::builder()
        .timeout(Duration::from_secs(TIMEOUT_SECONDS))
        .build()
        .map_err(|e| ApiError {
            message: format!("Failed to create HTTP client: {}", e),
        })?;

    let req = RegisterRequest { email, auk_hash };

    let response = client
        .post(format!("{}/auth/register", API_BASE_URL))
        .json(&req)
        .send()
        .map_err(|e| ApiError {
            message: format!("Registration request failed: {}", e),
        })?;

    if !response.status().is_success() {
        let error_text = response.text().unwrap_or_else(|_| "Unknown error".to_string());
        return Err(ApiError {
            message: format!("Registration failed: {}", error_text),
        });
    }

    response.json::<RegisterResponse>().map_err(|e| ApiError {
        message: format!("Failed to parse response: {}", e),
    })
}

/// Login user
#[tauri::command]
pub fn login_user(email: String, auk: String) -> Result<LoginResponse, ApiError> {
    let client = Client::builder()
        .timeout(Duration::from_secs(TIMEOUT_SECONDS))
        .build()
        .map_err(|e| ApiError {
            message: format!("Failed to create HTTP client: {}", e),
        })?;

    let req = LoginRequest { email, auk };

    let response = client
        .post(format!("{}/auth/login", API_BASE_URL))
        .json(&req)
        .send()
        .map_err(|e| ApiError {
            message: format!("Login request failed: {}", e),
        })?;

    if !response.status().is_success() {
        let error_text = response.text().unwrap_or_else(|_| "Unknown error".to_string());
        return Err(ApiError {
            message: format!("Login failed: {}", error_text),
        });
    }

    response.json::<LoginResponse>().map_err(|e| ApiError {
        message: format!("Failed to parse response: {}", e),
    })
}

/// Refresh access token
#[tauri::command]
pub fn refresh_token(refresh_token: String) -> Result<LoginResponse, ApiError> {
    let client = Client::builder()
        .timeout(Duration::from_secs(TIMEOUT_SECONDS))
        .build()
        .map_err(|e| ApiError {
            message: format!("Failed to create HTTP client: {}", e),
        })?;

    #[derive(Serialize)]
    struct RefreshRequest {
        refresh_token: String,
    }

    let req = RefreshRequest { refresh_token };

    let response = client
        .post(format!("{}/auth/refresh", API_BASE_URL))
        .json(&req)
        .send()
        .map_err(|e| ApiError {
            message: format!("Refresh request failed: {}", e),
        })?;

    if !response.status().is_success() {
        let error_text = response.text().unwrap_or_else(|_| "Unknown error".to_string());
        return Err(ApiError {
            message: format!("Token refresh failed: {}", error_text),
        });
    }

    response.json::<LoginResponse>().map_err(|e| ApiError {
        message: format!("Failed to parse response: {}", e),
    })
}

/// Get user's vault
#[tauri::command]
pub fn get_vault(access_token: String) -> Result<VaultResponse, ApiError> {
    let client = Client::builder()
        .timeout(Duration::from_secs(TIMEOUT_SECONDS))
        .build()
        .map_err(|e| ApiError {
            message: format!("Failed to create HTTP client: {}", e),
        })?;

    let response = client
        .get(format!("{}/vault", API_BASE_URL))
        .header("Authorization", format!("Bearer {}", access_token))
        .send()
        .map_err(|e| ApiError {
            message: format!("Get vault request failed: {}", e),
        })?;

    if !response.status().is_success() {
        let error_text = response.text().unwrap_or_else(|_| "Unknown error".to_string());
        return Err(ApiError {
            message: format!("Get vault failed: {}", error_text),
        });
    }

    response.json::<VaultResponse>().map_err(|e| ApiError {
        message: format!("Failed to parse response: {}", e),
    })
}

/// Update user's vault
#[tauri::command]
pub fn update_vault(
    access_token: String,
    encrypted_blob: String,
    version: i32,
) -> Result<UpdateVaultResponse, ApiError> {
    let client = Client::builder()
        .timeout(Duration::from_secs(TIMEOUT_SECONDS))
        .build()
        .map_err(|e| ApiError {
            message: format!("Failed to create HTTP client: {}", e),
        })?;

    let req = UpdateVaultRequest {
        encrypted_blob,
        version,
    };

    let response = client
        .put(format!("{}/vault", API_BASE_URL))
        .header("Authorization", format!("Bearer {}", access_token))
        .json(&req)
        .send()
        .map_err(|e| ApiError {
            message: format!("Update vault request failed: {}", e),
        })?;

    if !response.status().is_success() {
        let error_text = response.text().unwrap_or_else(|_| "Unknown error".to_string());
        return Err(ApiError {
            message: format!("Update vault failed: {}", error_text),
        });
    }

    response.json::<UpdateVaultResponse>().map_err(|e| ApiError {
        message: format!("Failed to parse response: {}", e),
    })
}
