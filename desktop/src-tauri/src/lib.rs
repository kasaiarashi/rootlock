mod crypto;
mod commands;
mod api;
mod biometric;

use commands::*;

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .plugin(tauri_plugin_store::Builder::default().build())
        .invoke_handler(tauri::generate_handler![
            // Crypto commands
            generate_secret_key,
            derive_master_key,
            derive_account_unlock_key,
            derive_vault_encryption_key,
            encrypt_data,
            decrypt_data,
            // API commands
            register_user,
            login_user,
            refresh_token,
            get_vault,
            update_vault,
            // Biometric commands (macOS)
            #[cfg(target_os = "macos")]
            biometric::macos::is_biometric_available,
            #[cfg(target_os = "macos")]
            biometric::macos::authenticate_biometric,
            #[cfg(target_os = "macos")]
            biometric::macos::store_biometric_credential,
            #[cfg(target_os = "macos")]
            biometric::macos::get_biometric_credential,
            #[cfg(target_os = "macos")]
            biometric::macos::delete_biometric_credential,
            // Biometric commands (fallback)
            #[cfg(not(target_os = "macos"))]
            biometric::fallback::is_biometric_available,
            #[cfg(not(target_os = "macos"))]
            biometric::fallback::authenticate_biometric,
            #[cfg(not(target_os = "macos"))]
            biometric::fallback::store_biometric_credential,
            #[cfg(not(target_os = "macos"))]
            biometric::fallback::get_biometric_credential,
            #[cfg(not(target_os = "macos"))]
            biometric::fallback::delete_biometric_credential,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
