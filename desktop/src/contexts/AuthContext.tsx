import React, { createContext, useContext, useState, useEffect } from 'react';
import { Store } from '@tauri-apps/plugin-store';
import { AuthState, User, AuthTokens } from '../types';
import * as api from '../api/tauri';

interface AuthContextType extends AuthState {
  login: (email: string, masterPassword: string, secretKey: string) => Promise<void>;
  unlock: (masterPassword: string) => Promise<void>;
  register: (email: string, masterPassword: string) => Promise<{ secretKey: string; userId: string }>;
  logout: () => Promise<void>;
  setMasterEncryptionKey: (key: string | null) => void;
  setTokens: (tokens: AuthTokens | null) => void;
  hasStoredSession: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [tokens, setTokens] = useState<AuthTokens | null>(null);
  const [masterEncryptionKey, setMasterEncryptionKey] = useState<string | null>(null);
  const [store, setStore] = useState<Store | null>(null);

  // Initialize store
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
    } catch (error) {
      console.error('Unlock failed:', error);
      // Make sure MEK stays null if validation fails
      setMasterEncryptionKey(null);
      throw error;
    }
  };

  const logout = async (): Promise<void> => {
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
        login,
        unlock,
        register,
        logout,
        setMasterEncryptionKey,
        setTokens,
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
