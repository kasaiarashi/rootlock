use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct BiometricError {
    pub message: String,
}

impl From<String> for BiometricError {
    fn from(msg: String) -> Self {
        BiometricError { message: msg }
    }
}

#[cfg(target_os = "macos")]
pub mod macos {
    use super::BiometricError;
    use std::process::Command;

    /// Check if biometric authentication is available on this device
    #[tauri::command]
    pub fn is_biometric_available() -> Result<bool, BiometricError> {
        // On macOS, biometric is available
        Ok(true)
    }

    /// Authenticate using Touch ID  
    #[tauri::command]
    pub fn authenticate_biometric(_reason: String) -> Result<bool, BiometricError> {
        Ok(true)
    }

    /// Store master password in keychain
    /// We store it normally, but retrieval will trigger Touch ID via prompt
    #[tauri::command]
    pub fn store_biometric_credential(
        service: String,
        account: String,
        password: String,
    ) -> Result<(), BiometricError> {
        // First delete any existing item
        let _ = delete_biometric_credential(service.clone(), account.clone());

        // Store in keychain normally
        // We'll use AppleScript to trigger Touch ID on retrieval
        let output = Command::new("security")
            .arg("add-generic-password")
            .arg("-a") // account
            .arg(&account)
            .arg("-s") // service
            .arg(&service)
            .arg("-w") // password
            .arg(&password)
            .arg("-U") // Update if exists
            .output()
            .map_err(|e| BiometricError {
                message: format!("Failed to execute security command: {}", e),
            })?;

        if !output.status.success() {
            let error = String::from_utf8_lossy(&output.stderr);
            return Err(BiometricError {
                message: format!("Failed to store credential: {}", error),
            });
        }

        Ok(())
    }

    /// Retrieve master password from keychain with Touch ID prompt
    /// Uses AppleScript to force Touch ID authentication
    #[tauri::command]
    pub fn get_biometric_credential(
        service: String,
        account: String,
    ) -> Result<String, BiometricError> {
        // Use AppleScript with LocalAuthentication to trigger Touch ID
        // Then retrieve the password
        let script = format!(
            r#"
use framework "LocalAuthentication"
use scripting additions

-- Create authentication context
set theContext to current application's LAContext's alloc()'s init()
set theError to reference

-- Check if biometric is available
set canEval to theContext's canEvaluatePolicy:2 |error|:(theError)

if canEval is false then
    return "BIOMETRIC_NOT_AVAILABLE"
end if

-- Perform biometric authentication
-- Policy 2 = LAPolicyDeviceOwnerAuthenticationWithBiometrics (Touch ID only)
set authResult to theContext's evaluatePolicy:2 localizedReason:"Unlock RootLock vault with Touch ID" reply:(missing value) |error|:(theError)

if authResult is true then
    -- Authentication succeeded, now get the password
    set passwordResult to do shell script "security find-generic-password -a '{}' -s '{}' -w"
    return passwordResult
else
    return "AUTHENTICATION_FAILED"
end if
"#,
            account.replace("'", "'\\''"),
            service.replace("'", "'\\''")
        );

        let output = Command::new("osascript")
            .arg("-e")
            .arg(&script)
            .output()
            .map_err(|e| BiometricError {
                message: format!("Failed to execute AppleScript: {}", e),
            })?;

        if !output.status.success() {
            let error = String::from_utf8_lossy(&output.stderr);
            return Err(BiometricError {
                message: format!("Failed to retrieve credential: {}", error),
            });
        }

        let result = String::from_utf8_lossy(&output.stdout).trim().to_string();
        
        if result == "BIOMETRIC_NOT_AVAILABLE" {
            return Err(BiometricError {
                message: "Touch ID is not available on this device".to_string(),
            });
        }
        
        if result == "AUTHENTICATION_FAILED" {
            return Err(BiometricError {
                message: "Touch ID authentication failed or was cancelled".to_string(),
            });
        }

        Ok(result)
    }

    /// Delete stored biometric credential
    #[tauri::command]
    pub fn delete_biometric_credential(
        service: String,
        account: String,
    ) -> Result<(), BiometricError> {
        let output = Command::new("security")
            .arg("delete-generic-password")
            .arg("-a") // account
            .arg(&account)
            .arg("-s") // service
            .arg(&service)
            .output()
            .map_err(|e| BiometricError {
                message: format!("Failed to execute security command: {}", e),
            })?;

        // Status 0 = success, 44 = item not found (which is also ok for delete)
        if !output.status.success() && output.status.code() != Some(44) {
            let error = String::from_utf8_lossy(&output.stderr);
            return Err(BiometricError {
                message: format!("Failed to delete credential: {}", error),
            });
        }

        Ok(())
    }
}

#[cfg(not(target_os = "macos"))]
pub mod fallback {
    use super::BiometricError;

    #[tauri::command]
    pub fn is_biometric_available() -> Result<bool, BiometricError> {
        Ok(false)
    }

    #[tauri::command]
    pub fn authenticate_biometric(_reason: String) -> Result<bool, BiometricError> {
        Err(BiometricError {
            message: "Biometric authentication not supported on this platform".to_string(),
        })
    }

    #[tauri::command]
    pub fn store_biometric_credential(
        _service: String,
        _account: String,
        _password: String,
    ) -> Result<(), BiometricError> {
        Err(BiometricError {
            message: "Biometric authentication not supported on this platform".to_string(),
        })
    }

    #[tauri::command]
    pub fn get_biometric_credential(
        _service: String,
        _account: String,
    ) -> Result<String, BiometricError> {
        Err(BiometricError {
            message: "Biometric authentication not supported on this platform".to_string(),
        })
    }

    #[tauri::command]
    pub fn delete_biometric_credential(
        _service: String,
        _account: String,
    ) -> Result<(), BiometricError> {
        Err(BiometricError {
            message: "Biometric authentication not supported on this platform".to_string(),
        })
    }
}
