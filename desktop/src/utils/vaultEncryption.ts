import * as api from '../api/tauri';
import { Vault, VaultItem, VaultItemData } from '../types';

/**
 * Encrypts vault items using per-item encryption keys derived from VEK
 */
export async function encryptVault(vault: Vault, vek: string): Promise<string> {
  const encryptedItems = await Promise.all(
    vault.items.map(async (item) => {
      // Derive per-item key using item ID
      const itemKey = await deriveItemKey(vek, item.id);
      
      // Serialize item data
      const plaintext = JSON.stringify(item);
      
      // Encrypt with item-specific key
      const ciphertext = await api.encryptData(plaintext, itemKey, item.id);
      
      return {
        id: item.id,
        ciphertext,
        createdAt: item.createdAt,
        updatedAt: item.updatedAt
      };
    })
  );
  
  const encryptedVault = {
    items: encryptedItems,
    version: vault.version
  };
  
  // Encrypt entire vault blob with VEK
  return await api.encryptData(JSON.stringify(encryptedVault), vek, 'vault');
}

/**
 * Decrypts vault blob and all items
 */
export async function decryptVault(encryptedBlob: string, vek: string): Promise<Vault> {
  // Decrypt vault blob
  const vaultJson = await api.decryptData(encryptedBlob, vek, 'vault');
  const encryptedVault = JSON.parse(vaultJson);
  
  // Decrypt individual items
  const items = await Promise.all(
    (encryptedVault.items || []).map(async (encItem: any) => {
      try {
        // Derive per-item key
        const itemKey = await deriveItemKey(vek, encItem.id);
        
        // Decrypt item
        const itemJson = await api.decryptData(encItem.ciphertext, itemKey, encItem.id);
        const item = JSON.parse(itemJson);
        
        return item as VaultItem;
      } catch (error) {
        console.error(`Failed to decrypt item ${encItem.id}:`, error);
        return null;
      }
    })
  );
  
  return {
    items: items.filter((item): item is VaultItem => item !== null),
    version: encryptedVault.version || 1
  };
}

/**
 * Derives a per-item encryption key from VEK and item ID using HKDF
 */
async function deriveItemKey(vek: string, itemId: string): Promise<string> {
  // Convert VEK from hex to bytes
  const vekBytes = hexToBytes(vek);
  
  // Use HKDF to derive item-specific key
  // Format: HKDF(vek, salt=itemId, info="item_encryption")
  const encoder = new TextEncoder();
  const salt = encoder.encode(itemId);
  const info = encoder.encode('item_encryption');
  
  // Import VEK as key material
  const keyMaterial = await crypto.subtle.importKey(
    'raw',
    vekBytes,
    'HKDF',
    false,
    ['deriveBits']
  );
  
  // Derive 256-bit key
  const derivedBits = await crypto.subtle.deriveBits(
    {
      name: 'HKDF',
      hash: 'SHA-256',
      salt: salt,
      info: info
    },
    keyMaterial,
    256
  );
  
  // Convert to hex string
  return bytesToHex(new Uint8Array(derivedBits));
}

/**
 * Hex string to byte array
 */
function hexToBytes(hex: string): Uint8Array {
  const bytes = new Uint8Array(hex.length / 2);
  for (let i = 0; i < hex.length; i += 2) {
    bytes[i / 2] = parseInt(hex.substring(i, i + 2), 16);
  }
  return bytes;
}

/**
 * Byte array to hex string
 */
function bytesToHex(bytes: Uint8Array): string {
  return Array.from(bytes)
    .map(b => b.toString(16).padStart(2, '0'))
    .join('');
}

/**
 * Creates an empty vault
 */
export function createEmptyVault(): Vault {
  return {
    items: [],
    version: 1
  };
}

/**
 * Adds an item to the vault
 */
export function addItemToVault(vault: Vault, itemData: VaultItemData): Vault {
  const newItem: VaultItem = {
    id: crypto.randomUUID(),
    data: itemData,
    createdAt: new Date().toISOString(),
    updatedAt: new Date().toISOString()
  };
  
  return {
    items: [...vault.items, newItem],
    version: vault.version + 1
  };
}

/**
 * Updates an existing item in the vault
 */
export function updateItemInVault(vault: Vault, itemId: string, itemData: VaultItemData): Vault {
  return {
    items: vault.items.map(item =>
      item.id === itemId
        ? { ...item, data: itemData, updatedAt: new Date().toISOString() }
        : item
    ),
    version: vault.version + 1
  };
}

/**
 * Deletes an item from the vault
 */
export function deleteItemFromVault(vault: Vault, itemId: string): Vault {
  return {
    items: vault.items.filter(item => item.id !== itemId),
    version: vault.version + 1
  };
}
