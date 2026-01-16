import React, { createContext, useContext, useState, useEffect } from 'react';
import { Store } from '@tauri-apps/plugin-store';
import { AuthState, User, AuthTokens } from '../types';
import * as api from '../api/tauri';

interface AuthContextType extends AuthState {
  login: (email: string, masterPassword: string, secretKey: string) => Promise<void>;
  unlock: (masterPassword: string) => Promise<void>;
  unlockWithBiometric: () => Promise<void>;
  register: (email: string, masterPassword: string) => Promise<{ secretKey: string; userId: string }>;
  logout: () => Promise<void>;
  setMasterEncryptionKey: (key: string | null) => void;
  setTokens: (tokens: AuthTokens | null) => void;
  hasStoredSession: boolean;
  biometricAvailable: boolean;
  enableBiometric: (masterPassword: string) => Promise<void>;
  disableBiometric: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [tokens, setTokens] = useState<AuthTokens | null>(null);
  const [masterEncryptionKey, setMasterEncryptionKey] = useState<string | null>(null);
  const [store, setStore] = useState<Store | null>(null);
  const [biometricAvailable, setBiometricAvailable] = useState(false);

  // Initialize store and check biometric availability
  useEffect(() => {
    async function initStore() {
      const s = await Store.load('auth.json');
      setStore(s);
      
      // Load stored auth data
      const storedUser = await s.get<User>('user');
      const storedTokens = await s.get<AuthTokens>('tokens');
      const storedSecretKey = await s.get<string>('secretKey');
      
      console.log('Loaded from store:', { 
        hasUser: !!storedUser, 
        hasTokens: !!storedTokens,
        hasSecretKey: !!storedSecretKey,
        userEmail: storedUser?.email
      });
      
      if (storedUser && storedTokens) {
        setUser(storedUser);
        setTokens(storedTokens);
      }

      // Check if biometric authentication is available
      const isBioAvailable = await api.isBiometricAvailable();
      setBiometricAvailable(isBioAvailable);
      console.log('Biometric available:', isBioAvailable);
    }
    initStore();
  }, []);

  const register = async (email: string, masterPassword: string): Promise<{ secretKey: string; userId: string }> => {
    try {
      // Generate secret key
      const secretKey = await api.generateSecretKey();
      
      // Derive MEK with email as salt (deterministic, doesn't require user ID yet)
      const mek = await api.deriveMasterKey(masterPassword, email, secretKey);
      
      // Derive AUK from MEK
      const auk = await api.deriveAccountUnlockKey(mek);
      
      // Register with backend (sends AUK hash)
      const registerResponse = await api.registerUser(email, auk);
      
      // After registration, login to get tokens
      const loginResponse = await api.loginUser(email, auk);
      
      // Store MEK in memory (not persisted for security)
      setMasterEncryptionKey(mek);
      
      // Store user info, tokens, and secret key
      setUser(loginResponse.user);
      setTokens(loginResponse.tokens);
      
      if (store) {
        await store.set('user', loginResponse.user);
        await store.set('tokens', loginResponse.tokens);
        await store.set('secretKey', secretKey); // Store secret key on device
        await store.save();
        console.log('Registration: Saved to store', { email: loginResponse.user.email, hasSecretKey: !!secretKey });
      }
      
      // Note: Secret key and user ID returned to user to save securely (as backup)
      return { secretKey, userId: registerResponse.user.id };
    } catch (error) {
      console.error('Registration failed:', error);
      throw error;
    }
  };

  const login = async (email: string, masterPassword: string, secretKey: string): Promise<void> => {
    try {
      // Derive MEK with email as salt (same as registration)
      const mek = await api.deriveMasterKey(masterPassword, email, secretKey);
      
      // Derive AUK from MEK
      const auk = await api.deriveAccountUnlockKey(mek);
      
      // Login with backend
      const response = await api.loginUser(email, auk);
      
      // Store auth data
      setUser(response.user);
      setTokens(response.tokens);
      setMasterEncryptionKey(mek);
      
      if (store) {
        await store.set('user', response.user);
        await store.set('tokens', response.tokens);
        await store.set('secretKey', secretKey); // Store secret key on device
        await store.save();
        console.log('Login: Saved to store', { email: response.user.email, hasSecretKey: !!secretKey });
      }
    } catch (error) {
      console.error('Login failed:', error);
      throw error;
    }
  };

  const unlock = async (masterPassword: string): Promise<void> => {
    try {
      if (!user || !store || !tokens) {
        throw new Error('No stored session found');
      }

      // Get stored secret key from device
      const secretKey = await store.get<string>('secretKey');
      if (!secretKey) {
        throw new Error('Secret key not found. Please login again.');
      }

      // Derive MEK with email as salt
      const mek = await api.deriveMasterKey(masterPassword, user.email, secretKey);
      
      // Try to decrypt vault to validate password is correct
      let vaultResponse;
      let currentAccessToken = tokens.access_token;
      
      try {
        vaultResponse = await api.getVault(currentAccessToken);
      } catch (error: any) {
        console.log('[Unlock] Vault access failed, checking if token expired...', error);
        
        // Check if it's a token error
        const errorStr = JSON.stringify(error).toLowerCase();
        const isTokenError = 
          (error.message && error.message.toLowerCase().includes('invalid or expired token')) ||
          errorStr.includes('invalid or expired token') ||
          errorStr.includes('unauthorized');
        
        if (isTokenError) {
          console.log('[Unlock] Token expired, attempting refresh...');
          try {
            // Attempt to refresh the token
            const refreshResponse = await api.refreshToken(tokens.refresh_token);
            console.log('[Unlock] Token refresh successful');
            
            // Update tokens in state and storage
            setTokens(refreshResponse.tokens);
            await store.set('tokens', refreshResponse.tokens);
            await store.save();
            
            // Retry with new token
            currentAccessToken = refreshResponse.tokens.access_token;
            vaultResponse = await api.getVault(currentAccessToken);
            console.log('[Unlock] Vault access successful with refreshed token');
          } catch (refreshError) {
            console.error('[Unlock] Token refresh failed:', refreshError);
            throw new Error('Session expired. Please login again.');
          }
        } else {
          // Not a token error, re-throw
          throw error;
        }
      }
      
      if (vaultResponse.encrypted_blob) {
        // Derive VEK and attempt decryption
        const vek = await api.deriveVaultEncryptionKey(mek);
        
        try {
          // This will throw if password is wrong
          await api.decryptData(vaultResponse.encrypted_blob, vek, 'vault');
        } catch (decryptError) {
          throw new Error('Invalid master password');
        }
      }
      
      // Only set MEK after successful validation
      setMasterEncryptionKey(mek);

      // Try to enable biometric unlock if available and not already enabled
      if (biometricAvailable && user) {
        try {
          const service = 'com.rootlock.vault';
          const account = user.email;
          
          // Check if credential already exists
          try {
            await api.getBiometricCredential(service, account);
            // Credential exists, no need to store again
          } catch {
            // Credential doesn't exist, store it
            await enableBiometric(masterPassword);
            console.log('Biometric unlock auto-enabled');
          }
        } catch (error) {
          console.log('Failed to auto-enable biometric:', error);
          // Don't throw - this is optional
        }
      }
    } catch (error) {
      console.error('Unlock failed:', error);
      // Make sure MEK stays null if validation fails
      setMasterEncryptionKey(null);
      throw error;
    }
  };

  const unlockWithBiometric = async (): Promise<void> => {
    console.log('[Biometric Unlock] Starting biometric unlock flow...');
    
    try {
      if (!user || !store || !tokens) {
        console.error('[Biometric Unlock] Missing required data:', { 
          hasUser: !!user, 
          hasStore: !!store, 
          hasTokens: !!tokens 
        });
        throw new Error('No stored session found');
      }

      console.log('[Biometric Unlock] User:', user.email);
      console.log('[Biometric Unlock] Biometric available:', biometricAvailable);

      if (!biometricAvailable) {
        throw new Error('Biometric authentication not available');
      }

      // Get stored secret key from device
      console.log('[Biometric Unlock] Retrieving secret key from store...');
      const secretKey = await store.get<string>('secretKey');
      if (!secretKey) {
        throw new Error('Secret key not found. Please login again.');
      }
      console.log('[Biometric Unlock] Secret key retrieved');

      // Retrieve master password from keychain
      // This will trigger the Touch ID prompt on macOS
      const service = 'com.rootlock.vault';
      const account = user.email;
      
      console.log('[Biometric Unlock] Retrieving master password from keychain...');
      console.log('[Biometric Unlock] Service:', service);
      console.log('[Biometric Unlock] Account:', account);
      console.log('[Biometric Unlock] This should trigger Touch ID prompt...');
      
      let masterPassword: string;
      try {
        masterPassword = await api.getBiometricCredential(service, account);
        console.log('[Biometric Unlock] Master password retrieved from keychain successfully');
      } catch (error) {
        console.error('[Biometric Unlock] Failed to get credential:', error);
        console.error('[Biometric Unlock] Error details:', JSON.stringify(error, null, 2));
        throw new Error('Master password not stored. Please unlock with password first.');
      }

      // Derive MEK with email as salt
      const mek = await api.deriveMasterKey(masterPassword, user.email, secretKey);
      
      // Try to decrypt vault to validate password is correct
      let vaultResponse;
      let currentAccessToken = tokens.access_token;
      
      try {
        vaultResponse = await api.getVault(currentAccessToken);
      } catch (error: any) {
        console.log('[Biometric Unlock] Vault access failed, checking if token expired...', error);
        
        // Check if it's a token error
        const errorStr = JSON.stringify(error).toLowerCase();
        const isTokenError = 
          (error.message && error.message.toLowerCase().includes('invalid or expired token')) ||
          errorStr.includes('invalid or expired token') ||
          errorStr.includes('unauthorized');
        
        if (isTokenError) {
          console.log('[Biometric Unlock] Token expired, attempting refresh...');
          try {
            const refreshResponse = await api.refreshToken(tokens.refresh_token);
            console.log('[Biometric Unlock] Token refresh successful');
            
            setTokens(refreshResponse.tokens);
            await store.set('tokens', refreshResponse.tokens);
            await store.save();
            
            currentAccessToken = refreshResponse.tokens.access_token;
            vaultResponse = await api.getVault(currentAccessToken);
            console.log('[Biometric Unlock] Vault access successful with refreshed token');
          } catch (refreshError) {
            console.error('[Biometric Unlock] Token refresh failed:', refreshError);
            throw new Error('Session expired. Please login again.');
          }
        } else {
          throw error;
        }
      }
      
      if (vaultResponse.encrypted_blob) {
        const vek = await api.deriveVaultEncryptionKey(mek);
        
        try {
          await api.decryptData(vaultResponse.encrypted_blob, vek, 'vault');
        } catch (decryptError) {
          throw new Error('Invalid master password in keychain');
        }
      }
      
      // Only set MEK after successful validation
      setMasterEncryptionKey(mek);
    } catch (error) {
      console.error('Biometric unlock failed:', error);
      setMasterEncryptionKey(null);
      throw error;
    }
  };

  const enableBiometric = async (masterPassword: string): Promise<void> => {
    console.log('[Enable Biometric] Starting...');
    
    if (!user) {
      throw new Error('No user logged in');
    }

    if (!biometricAvailable) {
      throw new Error('Biometric authentication not available on this device');
    }

    // Store master password in keychain
    // The keychain will handle Touch ID prompts automatically on macOS
    const service = 'com.rootlock.vault';
    const account = user.email;
    
    console.log('[Enable Biometric] Storing credential in keychain...');
    console.log('[Enable Biometric] Service:', service);
    console.log('[Enable Biometric] Account:', account);
    
    try {
      await api.storeBiometricCredential(service, account, masterPassword);
      console.log('[Enable Biometric] Success! Biometric unlock enabled for', user.email);
    } catch (error) {
      console.error('[Enable Biometric] Failed to store credential:', error);
      throw error;
    }
  };

  const disableBiometric = async (): Promise<void> => {
    if (!user) {
      throw new Error('No user logged in');
    }

    const service = 'com.rootlock.vault';
    const account = user.email;
    
    try {
      await api.deleteBiometricCredential(service, account);
      console.log('Biometric unlock disabled for', user.email);
    } catch (error) {
      console.error('Failed to disable biometric:', error);
      // Don't throw - credential might not exist
    }
  };

  const logout = async (): Promise<void> => {
    // Clean up biometric credentials
    await disableBiometric().catch(() => {});
    
    setUser(null);
    setTokens(null);
    setMasterEncryptionKey(null);
    
    if (store) {
      await store.clear();
      await store.save();
    }
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        tokens,
        masterEncryptionKey,
        isAuthenticated: !!user && !!tokens && !!masterEncryptionKey,
        hasStoredSession: !!user && !!tokens,
        biometricAvailable,
        login,
        unlock,
        unlockWithBiometric,
        register,
        logout,
        setMasterEncryptionKey,
        setTokens,
        enableBiometric,
        disableBiometric,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
