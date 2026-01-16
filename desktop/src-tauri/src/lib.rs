mod crypto;
mod commands;
mod api;

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
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
