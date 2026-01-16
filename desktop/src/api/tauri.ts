import { invoke } from '@tauri-apps/api/core';

// Crypto API
export async function generateSecretKey(): Promise<string> {
  return await invoke('generate_secret_key');
}

export async function deriveMasterKey(
  masterPassword: string,
  userId: string,
  secretKey: string
): Promise<string> {
  return await invoke('derive_master_key', {
    masterPassword,
    userId,
    secretKey,
  });
}

export async function deriveAccountUnlockKey(
  masterEncryptionKey: string
): Promise<string> {
  return await invoke('derive_account_unlock_key', {
    masterEncryptionKey,
  });
}

export async function deriveVaultEncryptionKey(
  masterEncryptionKey: string
): Promise<string> {
  return await invoke('derive_vault_encryption_key', {
    masterEncryptionKey,
  });
}

export async function encryptData(
  plaintext: string,
  key: string,
  associatedData?: string
): Promise<string> {
  return await invoke('encrypt_data', {
    plaintext,
    key,
    associatedData,
  });
}

export async function decryptData(
  ciphertext: string,
  key: string,
  associatedData?: string
): Promise<string> {
  return await invoke('decrypt_data', {
    ciphertext,
    key,
    associatedData,
  });
}

// Backend API
export interface RegisterResponse {
  message: string;
  user: {
    id: string;
    email: string;
    created_at: string;
  };
}

export interface LoginResponse {
  user: {
    id: string;
    email: string;
    created_at: string;
    last_login?: string;
  };
  tokens: {
    access_token: string;
    refresh_token: string;
    expires_in: number;
  };
}

export interface VaultResponse {
  id: string;
  encrypted_blob: string;
  version: number;
  last_modified: string;
}

export interface UpdateVaultResponse {
  id: string;
  version: number;
  last_modified: string;
  message: string;
}

export async function registerUser(
  email: string,
  aukHash: string
): Promise<RegisterResponse> {
  return await invoke('register_user', {
    email,
    aukHash,
  });
}

export async function loginUser(
  email: string,
  auk: string
): Promise<LoginResponse> {
  return await invoke('login_user', {
    email,
    auk,
  });
}

export async function getVault(accessToken: string): Promise<VaultResponse> {
  return await invoke('get_vault', {
    accessToken,
  });
}

export async function updateVault(
  accessToken: string,
  encryptedBlob: string,
  version: number
): Promise<UpdateVaultResponse> {
  return await invoke('update_vault', {
    accessToken,
    encryptedBlob,
    version,
  });
}
