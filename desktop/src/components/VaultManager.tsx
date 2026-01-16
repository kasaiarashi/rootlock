import { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { Vault, VaultItem, VaultItemData, LoginItem } from '../types';
import * as api from '../api/tauri';
import {
  encryptVault,
  decryptVault,
  createEmptyVault,
  addItemToVault,
  updateItemInVault,
  deleteItemFromVault
} from '../utils/vaultEncryption';
import AddItemModal from './AddItemModal';
import ItemDetailModal from './ItemDetailModal';
import './VaultManager.css';

export default function VaultManager() {
  const { user, tokens, masterEncryptionKey, logout } = useAuth();
  const [vault, setVault] = useState<Vault>(createEmptyVault());
  const [vaultVersion, setVaultVersion] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showAddModal, setShowAddModal] = useState(false);
  const [selectedItem, setSelectedItem] = useState<VaultItem | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [filterType, setFilterType] = useState<string>('all');
  const [showFavoritesOnly, setShowFavoritesOnly] = useState(false);

  useEffect(() => {
    loadVault();
  }, []);

  const loadVault = async () => {
    if (!tokens || !masterEncryptionKey) return;

    try {
      setLoading(true);
      setError('');
      const vaultResponse = await api.getVault(tokens.access_token);
      
      if (vaultResponse.encrypted_blob) {
        // Derive vault encryption key
        const vek = await api.deriveVaultEncryptionKey(masterEncryptionKey);
        
        // Decrypt vault using utility function
        const decryptedVault = await decryptVault(vaultResponse.encrypted_blob, vek);
        
        setVault(decryptedVault);
        setVaultVersion(vaultResponse.version);
      } else {
        // Empty vault
        setVault(createEmptyVault());
      }
    } catch (err: any) {
      console.error('Failed to load vault:', err);
      setError('Failed to load vault: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  const saveVault = async (updatedVault: Vault) => {
    if (!tokens || !masterEncryptionKey) return;

    try {
      setError('');
      // Derive vault encryption key
      const vek = await api.deriveVaultEncryptionKey(masterEncryptionKey);
      
      // Encrypt vault
      const encryptedBlob = await encryptVault(updatedVault, vek);
      
      // Save to backend
      const response = await api.updateVault(tokens.access_token, encryptedBlob, vaultVersion);
      
      setVault(updatedVault);
      setVaultVersion(response.version);
    } catch (err: any) {
      console.error('Failed to save vault:', err);
      setError('Failed to save vault: ' + err.message);
      throw err;
    }
  };

  const handleAddItem = async (itemData: VaultItemData) => {
    const updatedVault = addItemToVault(vault, itemData);
    await saveVault(updatedVault);
    setShowAddModal(false);
  };

  const handleUpdateItem = async (itemId: string, itemData: VaultItemData) => {
    const updatedVault = updateItemInVault(vault, itemId, itemData);
    await saveVault(updatedVault);
    setSelectedItem(null);
  };

  const handleDeleteItem = async (itemId: string) => {
    if (!confirm('Are you sure you want to delete this item?')) return;
    
    const updatedVault = deleteItemFromVault(vault, itemId);
    await saveVault(updatedVault);
    setSelectedItem(null);
  };

  const handleToggleFavorite = async (itemId: string) => {
    const item = vault.items.find(i => i.id === itemId);
    if (!item) return;

    const updatedData = {
      ...item.data,
      favorite: !item.data.favorite
    };

    const updatedVault = updateItemInVault(vault, itemId, updatedData);
    await saveVault(updatedVault);
  };

  const filteredItems = vault.items.filter(item => {
    // Filter by type
    if (filterType !== 'all' && item.data.type !== filterType) {
      return false;
    }

    // Filter by favorites
    if (showFavoritesOnly && !item.data.favorite) {
      return false;
    }

    // Search filter
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      const name = item.data.name?.toLowerCase() || '';
      
      if (name.includes(query)) return true;
      
      if (item.data.type === 'login') {
        const loginData = item.data as LoginItem;
        const username = loginData.username?.toLowerCase() || '';
        const url = loginData.url?.toLowerCase() || '';
        return username.includes(query) || url.includes(query);
      }
    }

    return searchQuery === '' || false;
  });

  const getItemIcon = (type: string) => {
    switch (type) {
      case 'login': return '🔑';
      case 'note': return '📝';
      case 'card': return '💳';
      case 'identity': return '👤';
      default: return '📄';
    }
  };

  const getItemSubtitle = (item: VaultItem): string => {
    switch (item.data.type) {
      case 'login':
        return (item.data as LoginItem).username || (item.data as LoginItem).url || '';
      case 'note':
        return 'Secure note';
      case 'card':
        return `Card ending in ${(item.data as any).cardNumber?.slice(-4) || '****'}`;
      case 'identity':
        return `${(item.data as any).firstName || ''} ${(item.data as any).lastName || ''}`.trim();
      default:
        return '';
    }
  };

  if (loading) {
    return (
      <div className="vault-manager">
        <div className="loading-state">
          <div className="spinner"></div>
          <p>Decrypting your vault...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="vault-manager">
      <header className="vault-header">
        <div className="header-left">
          <h1>🔐 RootLock</h1>
          <span className="item-count">{vault.items.length} items</span>
        </div>
        <div className="header-right">
          <span className="user-email">{user?.email}</span>
          <button onClick={logout} className="btn-logout">
            Logout
          </button>
        </div>
      </header>

      {error && (
        <div className="error-banner">
          <span>⚠️ {error}</span>
          <button onClick={() => setError('')}>×</button>
        </div>
      )}

      <div className="vault-toolbar">
        <div className="search-box">
          <span className="search-icon">🔍</span>
          <input
            type="text"
            placeholder="Search your vault..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="search-input"
          />
        </div>

        <div className="filter-controls">
          <select
            value={filterType}
            onChange={(e) => setFilterType(e.target.value)}
            className="filter-select"
          >
            <option value="all">All Types</option>
            <option value="login">🔑 Logins</option>
            <option value="note">📝 Notes</option>
            <option value="card">💳 Cards</option>
            <option value="identity">👤 Identities</option>
          </select>

          <button
            className={`btn-filter ${showFavoritesOnly ? 'active' : ''}`}
            onClick={() => setShowFavoritesOnly(!showFavoritesOnly)}
          >
            ⭐ Favorites
          </button>
        </div>

        <button onClick={() => setShowAddModal(true)} className="btn-add">
          + Add Item
        </button>
      </div>

      <div className="vault-content">
        {filteredItems.length === 0 ? (
          <div className="empty-state">
            {vault.items.length === 0 ? (
              <>
                <div className="empty-icon">🔒</div>
                <h2>Your vault is empty</h2>
                <p>Add your first password, note, or card to get started!</p>
                <button onClick={() => setShowAddModal(true)} className="btn-primary">
                  Add Your First Item
                </button>
              </>
            ) : (
              <>
                <div className="empty-icon">🔍</div>
                <h2>No items found</h2>
                <p>Try adjusting your search or filters</p>
              </>
            )}
          </div>
        ) : (
          <div className="items-grid">
            {filteredItems.map((item) => (
              <div
                key={item.id}
                className="vault-item-card"
                onClick={() => setSelectedItem(item)}
              >
                <div className="item-header">
                  <span className="item-icon">{getItemIcon(item.data.type)}</span>
                  {item.data.favorite && <span className="favorite-badge">⭐</span>}
                </div>
                <h3 className="item-name">{item.data.name}</h3>
                <p className="item-subtitle">{getItemSubtitle(item)}</p>
                <div className="item-footer">
                  <span className="item-type">{item.data.type}</span>
                  <span className="item-date">
                    {new Date(item.updatedAt).toLocaleDateString()}
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="vault-footer">
        <p>🔒 End-to-end encrypted • Zero-knowledge • Your data, your control</p>
      </div>

      {showAddModal && (
        <AddItemModal
          onClose={() => setShowAddModal(false)}
          onSave={handleAddItem}
        />
      )}

      {selectedItem && (
        <ItemDetailModal
          item={selectedItem}
          onClose={() => setSelectedItem(null)}
          onUpdate={handleUpdateItem}
          onDelete={handleDeleteItem}
          onToggleFavorite={handleToggleFavorite}
        />
      )}
    </div>
  );
}
