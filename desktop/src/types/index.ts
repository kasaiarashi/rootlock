export interface User {
  id: string;
  email: string;
  created_at: string;
  last_login?: string;
}

export interface AuthTokens {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface AuthState {
  user: User | null;
  tokens: AuthTokens | null;
  isAuthenticated: boolean;
  masterEncryptionKey: string | null;
}

export interface VaultItem {
  id: string;
  type: 'login' | 'note' | 'card' | 'identity';
  name: string;
  username?: string;
  password?: string;
  url?: string;
  notes?: string;
  createdAt: string;
  updatedAt: string;
}

export interface Vault {
  items: VaultItem[];
  version: number;
}
