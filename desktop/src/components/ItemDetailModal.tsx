import { useState } from 'react';
import { VaultItem, VaultItemData, LoginItem, NoteItem, CardItem, IdentityItem } from '../types';
import './Modal.css';

interface ItemDetailModalProps {
  item: VaultItem;
  onClose: () => void;
  onUpdate: (itemId: string, itemData: VaultItemData) => Promise<void>;
  onDelete: (itemId: string) => Promise<void>;
  onToggleFavorite: (itemId: string) => Promise<void>;
}

export default function ItemDetailModal({
  item,
  onClose,
  onUpdate,
  onDelete,
  onToggleFavorite
}: ItemDetailModalProps) {
  const [isEditing, setIsEditing] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [showCardNumber, setShowCardNumber] = useState(false);
  const [showCVV, setShowCVV] = useState(false);

  // Form state - initialized with item data
  const [formData, setFormData] = useState<any>({ ...item.data });

  const handleCopy = async (text: string, label: string) => {
    try {
      await navigator.clipboard.writeText(text);
      // Could add a toast notification here
      alert(`${label} copied to clipboard!`);
    } catch (err) {
      alert('Failed to copy to clipboard');
    }
  };

  const handleUpdate = async () => {
    setError('');
    setLoading(true);

    try {
      await onUpdate(item.id, formData as VaultItemData);
      setIsEditing(false);
    } catch (err: any) {
      setError(err.message || 'Failed to update item');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async () => {
    setLoading(true);
    try {
      await onDelete(item.id);
    } catch (err: any) {
      setError(err.message || 'Failed to delete item');
      setLoading(false);
    }
  };

  const handleToggleFavorite = async () => {
    setLoading(true);
    try {
      await onToggleFavorite(item.id);
    } catch (err: any) {
      setError(err.message || 'Failed to update favorite');
    } finally {
      setLoading(false);
    }
  };

  const renderViewMode = () => {
    switch (item.data.type) {
      case 'login':
        const loginData = item.data as LoginItem;
        return (
          <div className="item-fields">
            <div className="field-group">
              <label>Username</label>
              <div className="field-value">
                <span>{loginData.username}</span>
                <button onClick={() => handleCopy(loginData.username, 'Username')} className="btn-icon">
                  📋
                </button>
              </div>
            </div>

            <div className="field-group">
              <label>Password</label>
              <div className="field-value">
                <span className="password-value">
                  {showPassword ? loginData.password : '••••••••'}
                </span>
                <button onClick={() => setShowPassword(!showPassword)} className="btn-icon">
                  {showPassword ? '🙈' : '👁️'}
                </button>
                <button onClick={() => handleCopy(loginData.password, 'Password')} className="btn-icon">
                  📋
                </button>
              </div>
            </div>

            {loginData.url && (
              <div className="field-group">
                <label>Website</label>
                <div className="field-value">
                  <a href={loginData.url} target="_blank" rel="noopener noreferrer">
                    {loginData.url}
                  </a>
                </div>
              </div>
            )}

            {loginData.notes && (
              <div className="field-group">
                <label>Notes</label>
                <div className="field-value">
                  <p>{loginData.notes}</p>
                </div>
              </div>
            )}
          </div>
        );

      case 'note':
        const noteData = item.data as NoteItem;
        return (
          <div className="item-fields">
            <div className="field-group">
              <label>Content</label>
              <div className="field-value">
                <pre className="note-content">{noteData.content}</pre>
              </div>
            </div>
          </div>
        );

      case 'card':
        const cardData = item.data as CardItem;
        return (
          <div className="item-fields">
            <div className="field-group">
              <label>Cardholder Name</label>
              <div className="field-value">
                <span>{cardData.cardholderName}</span>
                <button onClick={() => handleCopy(cardData.cardholderName, 'Name')} className="btn-icon">
                  📋
                </button>
              </div>
            </div>

            <div className="field-group">
              <label>Card Number</label>
              <div className="field-value">
                <span className="password-value">
                  {showCardNumber ? cardData.cardNumber : `•••• •••• •••• ${cardData.cardNumber.slice(-4)}`}
                </span>
                <button onClick={() => setShowCardNumber(!showCardNumber)} className="btn-icon">
                  {showCardNumber ? '🙈' : '👁️'}
                </button>
                <button onClick={() => handleCopy(cardData.cardNumber, 'Card Number')} className="btn-icon">
                  📋
                </button>
              </div>
            </div>

            <div className="field-row">
              <div className="field-group">
                <label>Expiry</label>
                <div className="field-value">
                  <span>{cardData.expiryMonth}/{cardData.expiryYear}</span>
                </div>
              </div>

              <div className="field-group">
                <label>CVV</label>
                <div className="field-value">
                  <span className="password-value">
                    {showCVV ? cardData.cvv : '•••'}
                  </span>
                  <button onClick={() => setShowCVV(!showCVV)} className="btn-icon">
                    {showCVV ? '🙈' : '👁️'}
                  </button>
                  <button onClick={() => handleCopy(cardData.cvv, 'CVV')} className="btn-icon">
                    📋
                  </button>
                </div>
              </div>
            </div>

            {cardData.notes && (
              <div className="field-group">
                <label>Notes</label>
                <div className="field-value">
                  <p>{cardData.notes}</p>
                </div>
              </div>
            )}
          </div>
        );

      case 'identity':
        const identityData = item.data as IdentityItem;
        return (
          <div className="item-fields">
            <div className="field-row">
              <div className="field-group">
                <label>First Name</label>
                <div className="field-value">
                  <span>{identityData.firstName}</span>
                </div>
              </div>

              <div className="field-group">
                <label>Last Name</label>
                <div className="field-value">
                  <span>{identityData.lastName}</span>
                </div>
              </div>
            </div>

            {identityData.email && (
              <div className="field-group">
                <label>Email</label>
                <div className="field-value">
                  <span>{identityData.email}</span>
                  <button onClick={() => handleCopy(identityData.email!, 'Email')} className="btn-icon">
                    📋
                  </button>
                </div>
              </div>
            )}

            {identityData.phone && (
              <div className="field-group">
                <label>Phone</label>
                <div className="field-value">
                  <span>{identityData.phone}</span>
                  <button onClick={() => handleCopy(identityData.phone!, 'Phone')} className="btn-icon">
                    📋
                  </button>
                </div>
              </div>
            )}

            {identityData.address && (
              <div className="field-group">
                <label>Address</label>
                <div className="field-value">
                  <p>{identityData.address}</p>
                </div>
              </div>
            )}

            {identityData.notes && (
              <div className="field-group">
                <label>Notes</label>
                <div className="field-value">
                  <p>{identityData.notes}</p>
                </div>
              </div>
            )}
          </div>
        );
    }
  };

  const renderEditMode = () => {
    switch (item.data.type) {
      case 'login':
        return (
          <>
            <div className="form-group">
              <label>Name *</label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                required
              />
            </div>

            <div className="form-group">
              <label>Username *</label>
              <input
                type="text"
                value={formData.username}
                onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                required
              />
            </div>

            <div className="form-group">
              <label>Password *</label>
              <input
                type="text"
                value={formData.password}
                onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                required
              />
            </div>

            <div className="form-group">
              <label>Website URL</label>
              <input
                type="url"
                value={formData.url || ''}
                onChange={(e) => setFormData({ ...formData, url: e.target.value })}
              />
            </div>

            <div className="form-group">
              <label>Notes</label>
              <textarea
                value={formData.notes || ''}
                onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
                rows={3}
              />
            </div>
          </>
        );

      case 'note':
        return (
          <>
            <div className="form-group">
              <label>Name *</label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                required
              />
            </div>

            <div className="form-group">
              <label>Content *</label>
              <textarea
                value={formData.content}
                onChange={(e) => setFormData({ ...formData, content: e.target.value })}
                rows={10}
                required
              />
            </div>
          </>
        );

      case 'card':
        return (
          <>
            <div className="form-group">
              <label>Name *</label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                required
              />
            </div>

            <div className="form-group">
              <label>Cardholder Name *</label>
              <input
                type="text"
                value={formData.cardholderName}
                onChange={(e) => setFormData({ ...formData, cardholderName: e.target.value })}
                required
              />
            </div>

            <div className="form-group">
              <label>Card Number *</label>
              <input
                type="text"
                value={formData.cardNumber}
                onChange={(e) => setFormData({ ...formData, cardNumber: e.target.value })}
                maxLength={19}
                required
              />
            </div>

            <div className="form-row">
              <div className="form-group">
                <label>Expiry Month *</label>
                <input
                  type="text"
                  value={formData.expiryMonth}
                  onChange={(e) => setFormData({ ...formData, expiryMonth: e.target.value })}
                  maxLength={2}
                  required
                />
              </div>

              <div className="form-group">
                <label>Expiry Year *</label>
                <input
                  type="text"
                  value={formData.expiryYear}
                  onChange={(e) => setFormData({ ...formData, expiryYear: e.target.value })}
                  maxLength={4}
                  required
                />
              </div>

              <div className="form-group">
                <label>CVV *</label>
                <input
                  type="text"
                  value={formData.cvv}
                  onChange={(e) => setFormData({ ...formData, cvv: e.target.value })}
                  maxLength={4}
                  required
                />
              </div>
            </div>

            <div className="form-group">
              <label>Notes</label>
              <textarea
                value={formData.notes || ''}
                onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
                rows={3}
              />
            </div>
          </>
        );

      case 'identity':
        return (
          <>
            <div className="form-group">
              <label>Name *</label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                required
              />
            </div>

            <div className="form-row">
              <div className="form-group">
                <label>First Name *</label>
                <input
                  type="text"
                  value={formData.firstName}
                  onChange={(e) => setFormData({ ...formData, firstName: e.target.value })}
                  required
                />
              </div>

              <div className="form-group">
                <label>Last Name *</label>
                <input
                  type="text"
                  value={formData.lastName}
                  onChange={(e) => setFormData({ ...formData, lastName: e.target.value })}
                  required
                />
              </div>
            </div>

            <div className="form-group">
              <label>Email</label>
              <input
                type="email"
                value={formData.email || ''}
                onChange={(e) => setFormData({ ...formData, email: e.target.value })}
              />
            </div>

            <div className="form-group">
              <label>Phone</label>
              <input
                type="tel"
                value={formData.phone || ''}
                onChange={(e) => setFormData({ ...formData, phone: e.target.value })}
              />
            </div>

            <div className="form-group">
              <label>Address</label>
              <textarea
                value={formData.address || ''}
                onChange={(e) => setFormData({ ...formData, address: e.target.value })}
                rows={3}
              />
            </div>

            <div className="form-group">
              <label>Notes</label>
              <textarea
                value={formData.notes || ''}
                onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
                rows={3}
              />
            </div>
          </>
        );
    }
  };

  const getTypeIcon = () => {
    switch (item.data.type) {
      case 'login': return '🔑';
      case 'note': return '📝';
      case 'card': return '💳';
      case 'identity': return '👤';
    }
  };

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content large" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <div className="header-left">
            <span className="item-type-icon">{getTypeIcon()}</span>
            <h2>{item.data.name}</h2>
            <button
              onClick={handleToggleFavorite}
              className="btn-favorite"
              disabled={loading}
            >
              {item.data.favorite ? '⭐' : '☆'}
            </button>
          </div>
          <button className="modal-close" onClick={onClose}>×</button>
        </div>

        <div className="modal-body">
          {error && <div className="error-message">{error}</div>}

          {isEditing ? (
            <form onSubmit={(e) => { e.preventDefault(); handleUpdate(); }}>
              {renderEditMode()}
              <div className="modal-footer">
                <button
                  type="button"
                  onClick={() => {
                    setIsEditing(false);
                    setFormData({ ...item.data });
                  }}
                  className="btn-secondary"
                  disabled={loading}
                >
                  Cancel
                </button>
                <button type="submit" className="btn-primary" disabled={loading}>
                  {loading ? 'Saving...' : 'Save Changes'}
                </button>
              </div>
            </form>
          ) : (
            <>
              {renderViewMode()}
              <div className="item-metadata">
                <p>Created: {new Date(item.createdAt).toLocaleString()}</p>
                <p>Updated: {new Date(item.updatedAt).toLocaleString()}</p>
              </div>
              <div className="modal-footer">
                <button onClick={handleDelete} className="btn-danger" disabled={loading}>
                  {loading ? 'Deleting...' : 'Delete'}
                </button>
                <div className="footer-right">
                  <button onClick={() => setIsEditing(true)} className="btn-secondary">
                    Edit
                  </button>
                </div>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
