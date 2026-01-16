import React, { createContext, useContext, useState, useEffect } from 'react';
import { Store } from '@tauri-apps/plugin-store';
import { AuthState, User, AuthTokens } from '../types';
import * as api from '../api/tauri';

interface AuthContextType extends AuthState {
  login: (email: string, masterPassword: string, secretKey: string) => Promise<void>;
  register: (email: string, masterPassword: string) => Promise<{ secretKey: string; userId: string }>;
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

  const register = async (email: string, masterPassword: string): Promise<{ secretKey: string; userId: string }> => {
    try {
      // Generate secret key
      const secretKey = await api.generateSecretKey();
      
      // Derive MEK with email as salt (deterministic, doesn't require user ID yet)
      const mek = await api.deriveMasterKey(masterPassword, email, secretKey);
      
      // Derive AUK from MEK
      const auk = await api.deriveAccountUnlockKey(mek);
      
      // Register with backend (sends AUK hash)
      const response = await api.registerUser(email, auk);
      
      // Store MEK in memory (not persisted for security)
      setMasterEncryptionKey(mek);
      
      // Store user info
      setUser(response.user);
      
      // Note: Secret key and user ID returned to user to save securely
      return { secretKey, userId: response.user.id };
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
        isAuthenticated: !!user && !!tokens && !!masterEncryptionKey,
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
