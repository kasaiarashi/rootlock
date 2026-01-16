import { useState, useEffect } from 'react';
import { useAuth } from '../contexts/AuthContext';
import { Vault, VaultItem, VaultItemData } from '../types';
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
import './VaultManagerNew.css';

type CategoryType = 'all' | 'login' | 'note' | 'card' | 'identity' | 'favorites';

export default function VaultManagerNew() {
  const { user, tokens, masterEncryptionKey, logout } = useAuth();
  const [vault, setVault] = useState<Vault>(createEmptyVault());
  const [vaultVersion, setVaultVersion] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [selectedCategory, setSelectedCategory] = useState<CategoryType>('all');
  const [selectedItem, setSelectedItem] = useState<VaultItem | null>(null);
  const [showAddModal, setShowAddModal] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [isEditing, setIsEditing] = useState(false);
  const [editFormData, setEditFormData] = useState<VaultItemData | null>(null);

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
        const vek = await api.deriveVaultEncryptionKey(masterEncryptionKey);
        const decryptedVault = await decryptVault(vaultResponse.encrypted_blob, vek);
        
        setVault(decryptedVault);
        setVaultVersion(vaultResponse.version);
      } else {
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
      const vek = await api.deriveVaultEncryptionKey(masterEncryptionKey);
      const encryptedBlob = await encryptVault(updatedVault, vek);
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

  const handleUpdateItem = async () => {
    if (!selectedItem || !editFormData) return;
    
    const updatedVault = updateItemInVault(vault, selectedItem.id, editFormData);
    await saveVault(updatedVault);
    setSelectedItem({ ...selectedItem, data: editFormData });
    setIsEditing(false);
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
    
    if (selectedItem?.id === itemId) {
      setSelectedItem({ ...selectedItem, data: updatedData });
    }
  };

  const filteredItems = vault.items.filter(item => {
    // Category filter
    if (selectedCategory === 'favorites' && !item.data.favorite) return false;
    if (selectedCategory !== 'all' && selectedCategory !== 'favorites' && item.data.type !== selectedCategory) return false;

    // Search filter
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      const name = item.data.name?.toLowerCase() || '';
      if (name.includes(query)) return true;
      
      if (item.data.type === 'login') {
        const username = (item.data as any).username?.toLowerCase() || '';
        const url = (item.data as any).url?.toLowerCase() || '';
        return username.includes(query) || url.includes(query);
      }
      return false;
    }

    return true;
  });

  const getCategoryCount = (category: CategoryType): number => {
    if (category === 'all') return vault.items.length;
    if (category === 'favorites') return vault.items.filter(i => i.data.favorite).length;
    return vault.items.filter(i => i.data.type === category).length;
  };

  const getItemIcon = (type: string) => {
    switch (type) {
      case 'login': return 'key';
      case 'note': return 'file-text';
      case 'card': return 'credit-card';
      case 'identity': return 'user';
      default: return 'file';
    }
  };

  const getItemSubtitle = (item: VaultItem): string => {
    switch (item.data.type) {
      case 'login':
        return (item.data as any).username || (item.data as any).url || '';
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

  const startEditing = () => {
    if (selectedItem) {
      setEditFormData({ ...selectedItem.data });
      setIsEditing(true);
    }
  };

  const cancelEditing = () => {
    setIsEditing(false);
    setEditFormData(null);
  };

  if (loading) {
    return (
      <div className="vault-new-container">
        <div className="loading-screen">
          <div className="loading-spinner"></div>
          <p>Loading vault...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="vault-new-container">
      {/* Left Sidebar */}
      <aside className="vault-sidebar">
        <div className="sidebar-header">
          <div className="logo">RootLock</div>
          <button className="btn-add-new" onClick={() => setShowAddModal(true)}>
            New Item
          </button>
        </div>

        <nav className="sidebar-nav">
          <button
            className={`nav-item ${selectedCategory === 'all' ? 'active' : ''}`}
            onClick={() => setSelectedCategory('all')}
          >
            <span className="nav-icon">folder</span>
            <span className="nav-label">All Items</span>
            <span className="nav-count">{getCategoryCount('all')}</span>
          </button>

          <button
            className={`nav-item ${selectedCategory === 'favorites' ? 'active' : ''}`}
            onClick={() => setSelectedCategory('favorites')}
          >
            <span className="nav-icon">star</span>
            <span className="nav-label">Favorites</span>
            <span className="nav-count">{getCategoryCount('favorites')}</span>
          </button>

          <div className="nav-divider"></div>

          <div className="nav-section-title">Categories</div>

          <button
            className={`nav-item ${selectedCategory === 'login' ? 'active' : ''}`}
            onClick={() => setSelectedCategory('login')}
          >
            <span className="nav-icon">key</span>
            <span className="nav-label">Logins</span>
            <span className="nav-count">{getCategoryCount('login')}</span>
          </button>

          <button
            className={`nav-item ${selectedCategory === 'note' ? 'active' : ''}`}
            onClick={() => setSelectedCategory('note')}
          >
            <span className="nav-icon">file-text</span>
            <span className="nav-label">Notes</span>
            <span className="nav-count">{getCategoryCount('note')}</span>
          </button>

          <button
            className={`nav-item ${selectedCategory === 'card' ? 'active' : ''}`}
            onClick={() => setSelectedCategory('card')}
          >
            <span className="nav-icon">credit-card</span>
            <span className="nav-label">Cards</span>
            <span className="nav-count">{getCategoryCount('card')}</span>
          </button>

          <button
            className={`nav-item ${selectedCategory === 'identity' ? 'active' : ''}`}
            onClick={() => setSelectedCategory('identity')}
          >
            <span className="nav-icon">user</span>
            <span className="nav-label">Identities</span>
            <span className="nav-count">{getCategoryCount('identity')}</span>
          </button>
        </nav>

        <div className="sidebar-footer">
          <div className="vault-switcher">
            <div className="current-vault">
              <span className="vault-icon">database</span>
              <div className="vault-info">
                <div className="vault-name">Personal Vault</div>
                <div className="vault-email">{user?.email}</div>
              </div>
            </div>
          </div>
          <button className="btn-logout" onClick={logout}>
            Logout
          </button>
        </div>
      </aside>

      {/* Middle Panel - Item List */}
      <main className="vault-main">
        <div className="main-header">
          <div className="search-bar">
            <span className="search-icon">search</span>
            <input
              type="text"
              placeholder="Search vault..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>
        </div>

        {error && (
          <div className="error-banner">
            {error}
            <button onClick={() => setError('')}>×</button>
          </div>
        )}

        <div className="items-list">
          {filteredItems.length === 0 ? (
            <div className="empty-state">
              <div className="empty-icon">inbox</div>
              <h3>No items found</h3>
              {vault.items.length === 0 ? (
                <p>Get started by creating your first vault item</p>
              ) : (
                <p>Try adjusting your search or filters</p>
              )}
            </div>
          ) : (
            filteredItems.map(item => (
              <div
                key={item.id}
                className={`list-item ${selectedItem?.id === item.id ? 'selected' : ''}`}
                onClick={() => {
                  setSelectedItem(item);
                  setIsEditing(false);
                }}
              >
                <div className="item-icon-wrapper">
                  <span className="item-icon">{getItemIcon(item.data.type)}</span>
                </div>
                <div className="item-content">
                  <div className="item-name">{item.data.name}</div>
                  <div className="item-subtitle">{getItemSubtitle(item)}</div>
                </div>
                {item.data.favorite && <span className="favorite-indicator">star</span>}
              </div>
            ))
          )}
        </div>
      </main>

      {/* Right Panel - Item Details */}
      <aside className="vault-detail">
        {selectedItem ? (
          <div className="detail-container">
            <div className="detail-header">
              <div className="detail-title-row">
                <h2>{selectedItem.data.name}</h2>
                <button
                  className="btn-favorite-toggle"
                  onClick={() => handleToggleFavorite(selectedItem.id)}
                  title={selectedItem.data.favorite ? 'Remove from favorites' : 'Add to favorites'}
                >
                  {selectedItem.data.favorite ? 'star-filled' : 'star'}
                </button>
              </div>
              <div className="detail-type">{selectedItem.data.type}</div>
            </div>

            <div className="detail-actions">
              {!isEditing ? (
                <>
                  <button className="btn-action" onClick={startEditing}>
                    Edit
                  </button>
                  <button className="btn-action btn-danger" onClick={() => handleDeleteItem(selectedItem.id)}>
                    Delete
                  </button>
                </>
              ) : (
                <>
                  <button className="btn-action btn-primary" onClick={handleUpdateItem}>
                    Save Changes
                  </button>
                  <button className="btn-action" onClick={cancelEditing}>
                    Cancel
                  </button>
                </>
              )}
            </div>

            <div className="detail-content">
              {/* Render fields based on item type and edit mode */}
              {/* This will be expanded in the next section */}
              <p>Details panel content here...</p>
            </div>
          </div>
        ) : (
          <div className="detail-empty">
            <div className="empty-icon">select</div>
            <h3>No item selected</h3>
            <p>Select an item from the list to view its details</p>
          </div>
        )}
      </aside>

      {showAddModal && (
        <AddItemModal
          onClose={() => setShowAddModal(false)}
          onSave={handleAddItem}
        />
      )}
    </div>
  );
}
