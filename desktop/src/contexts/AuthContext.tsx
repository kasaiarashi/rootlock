import React, { createContext, useContext, useState, useEffect } from 'react';
import { Store } from '@tauri-apps/plugin-store';
import { AuthState, User, AuthTokens } from '../types';
import * as api from '../api/tauri';

interface AuthContextType extends AuthState {
  login: (email: string, masterPassword: string, secretKey: string) => Promise<void>;
  register: (email: string, masterPassword: string) => Promise<{ secretKey: string }>;
  logout: () => Promise<void>;
  setMasterEncryptionKey: (key: string | null) => void;
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
      
      if (storedUser && storedTokens) {
        setUser(storedUser);
        setTokens(storedTokens);
      }
    }
    initStore();
  }, []);

  const register = async (email: string, masterPassword: string): Promise<{ secretKey: string }> => {
    try {
      // Generate secret key
      const secretKey = await api.generateSecretKey();
      
      // Derive MEK
      const tempUserId = crypto.randomUUID(); // Temporary until we get real user ID
      const mek = await api.deriveMasterKey(masterPassword, tempUserId, secretKey);
      
      // Derive AUK
      const auk = await api.deriveAccountUnlockKey(mek);
      
      // Register with backend
      const response = await api.registerUser(email, auk);
      
      // Now derive MEK again with actual user ID
      const actualMek = await api.deriveMasterKey(masterPassword, response.user.id, secretKey);
      
      // Store MEK in memory (not persisted for security)
      setMasterEncryptionKey(actualMek);
      
      // Note: Secret key is returned to user to save securely
      return { secretKey };
    } catch (error) {
      console.error('Registration failed:', error);
      throw error;
    }
  };

  const login = async (email: string, masterPassword: string, secretKey: string): Promise<void> => {
    try {
      // Note: We need user ID to derive MEK, but we don't have it yet
      // So we'll do a two-step process:
      // 1. First login to get user ID
      // 2. Then derive MEK with correct user ID
      
      // For now, use temporary derivation
      const tempUserId = crypto.randomUUID();
      const tempMek = await api.deriveMasterKey(masterPassword, tempUserId, secretKey);
      const auk = await api.deriveAccountUnlockKey(tempMek);
      
      // Login with backend
      const response = await api.loginUser(email, auk);
      
      // Now derive actual MEK with real user ID
      const mek = await api.deriveMasterKey(masterPassword, response.user.id, secretKey);
      
      // Store auth data
      setUser(response.user);
      setTokens(response.tokens);
      setMasterEncryptionKey(mek);
      
      if (store) {
        await store.set('user', response.user);
        await store.set('tokens', response.tokens);
        await store.save();
      }
    } catch (error) {
      console.error('Login failed:', error);
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
        isAuthenticated: !!user && !!tokens,
        login,
        register,
        logout,
        setMasterEncryptionKey,
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
