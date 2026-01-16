import { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { VaultItem } from '../types';
import * as api from '../api/tauri';
import './Vault.css';

export default function Vault() {
  const { user, tokens, masterEncryptionKey, logout } = useAuth();
  const [items, setItems] = useState<VaultItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  useEffect(() => {
    loadVault();
  }, []);

  const loadVault = async () => {
    if (!tokens || !masterEncryptionKey) return;

    try {
      setLoading(true);
      const vaultResponse = await api.getVault(tokens.access_token);
      
      if (vaultResponse.encrypted_blob) {
        // Derive vault encryption key
        const vek = await api.deriveVaultEncryptionKey(masterEncryptionKey);
        
        // Decrypt vault
        const decryptedJson = await api.decryptData(vaultResponse.encrypted_blob, vek);
        const vault = JSON.parse(decryptedJson);
        
        setItems(vault.items || []);
      }
    } catch (err: any) {
      console.error('Failed to load vault:', err);
      setError('Failed to load vault');
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = async () => {
    await logout();
  };

  if (loading) {
    return (
      <div className="vault-container">
        <p>Loading vault...</p>
      </div>
    );
  }

  return (
    <div className="vault-container">
      <header className="vault-header">
        <h1>🔐 RootLock Vault</h1>
        <div className="user-info">
          <span>{user?.email}</span>
          <button onClick={handleLogout} className="btn-secondary">
            Logout
          </button>
        </div>
      </header>

      {error && <div className="error-message">{error}</div>}

      <div className="vault-content">
        {items.length === 0 ? (
          <div className="empty-state">
            <p>Your vault is empty</p>
            <p>Add your first password to get started!</p>
          </div>
        ) : (
          <div className="items-list">
            {items.map((item) => (
              <div key={item.id} className="vault-item">
                <div className="item-icon">
                  {item.data.type === 'login' && '🔑'}
                  {item.data.type === 'note' && '📝'}
                  {item.data.type === 'card' && '💳'}
                  {item.data.type === 'identity' && '👤'}
                </div>
                <div className="item-details">
                  <h3>{item.data.name}</h3>
                  {item.data.type === 'login' && (item.data as any).username && <p className="item-username">{(item.data as any).username}</p>}
                  {item.data.type === 'login' && (item.data as any).url && <p className="item-url">{(item.data as any).url}</p>}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="vault-footer">
        <p className="info-text">
          ✅ End-to-end encrypted • Zero-knowledge architecture
        </p>
      </div>
    </div>
  );
}
