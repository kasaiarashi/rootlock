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

export interface CustomField {
  id: string;
  label: string;
  value: string;
  type: 'text' | 'password' | 'email' | 'url' | 'number';
  hidden?: boolean; // For password-like fields
}

export interface LoginItem {
  type: 'login';
  name: string;
  username: string;
  password: string;
  url?: string;
  notes?: string;
  favorite?: boolean;
  tags?: string[];
  customFields?: CustomField[];
}

export interface NoteItem {
  type: 'note';
  name: string;
  content: string;
  favorite?: boolean;
  tags?: string[];
  customFields?: CustomField[];
}

export interface CardItem {
  type: 'card';
  name: string;
  cardholderName: string;
  cardNumber: string;
  expiryMonth: string;
  expiryYear: string;
  cvv: string;
  notes?: string;
  favorite?: boolean;
  tags?: string[];
  customFields?: CustomField[];
}

export interface IdentityItem {
  type: 'identity';
  name: string;
  firstName: string;
  lastName: string;
  email?: string;
  phone?: string;
  address?: string;
  notes?: string;
  favorite?: boolean;
  tags?: string[];
  customFields?: CustomField[];
}

export type VaultItemData = LoginItem | NoteItem | CardItem | IdentityItem;

export interface VaultItem {
  id: string;
  data: VaultItemData;
  createdAt: string;
  updatedAt: string;
}

export interface Vault {
  items: VaultItem[];
  version: number;
}
