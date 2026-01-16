import { useState } from 'react';
import { VaultItemData, LoginItem, NoteItem, CardItem, IdentityItem } from '../types';
import { generatePassword, PasswordOptions, calculatePasswordStrength } from '../utils/passwordGenerator';
import './Modal.css';

interface AddItemModalProps {
  onClose: () => void;
  onSave: (itemData: VaultItemData) => Promise<void>;
}

export default function AddItemModal({ onClose, onSave }: AddItemModalProps) {
  const [itemType, setItemType] = useState<'login' | 'note' | 'card' | 'identity'>('login');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [showPasswordGenerator, setShowPasswordGenerator] = useState(false);

  // Login fields
  const [loginName, setLoginName] = useState('');
  const [loginUsername, setLoginUsername] = useState('');
  const [loginPassword, setLoginPassword] = useState('');
  const [loginUrl, setLoginUrl] = useState('');
  const [loginNotes, setLoginNotes] = useState('');

  // Note fields
  const [noteName, setNoteName] = useState('');
  const [noteContent, setNoteContent] = useState('');

  // Card fields
  const [cardName, setCardName] = useState('');
  const [cardholderName, setCardholderName] = useState('');
  const [cardNumber, setCardNumber] = useState('');
  const [expiryMonth, setExpiryMonth] = useState('');
  const [expiryYear, setExpiryYear] = useState('');
  const [cvv, setCvv] = useState('');
  const [cardNotes, setCardNotes] = useState('');

  // Identity fields
  const [identityName, setIdentityName] = useState('');
  const [firstName, setFirstName] = useState('');
  const [lastName, setLastName] = useState('');
  const [email, setEmail] = useState('');
  const [phone, setPhone] = useState('');
  const [address, setAddress] = useState('');
  const [identityNotes, setIdentityNotes] = useState('');

  // Password generator
  const [passwordLength, setPasswordLength] = useState(16);
  const [includeUppercase, setIncludeUppercase] = useState(true);
  const [includeLowercase, setIncludeLowercase] = useState(true);
  const [includeNumbers, setIncludeNumbers] = useState(true);
  const [includeSymbols, setIncludeSymbols] = useState(true);

  const handleGeneratePassword = () => {
    try {
      const options: PasswordOptions = {
        length: passwordLength,
        uppercase: includeUppercase,
        lowercase: includeLowercase,
        numbers: includeNumbers,
        symbols: includeSymbols
      };
      const newPassword = generatePassword(options);
      setLoginPassword(newPassword);
    } catch (err: any) {
      setError(err.message);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      let itemData: VaultItemData;

      switch (itemType) {
        case 'login':
          if (!loginName || !loginUsername || !loginPassword) {
            throw new Error('Name, username, and password are required');
          }
          itemData = {
            type: 'login',
            name: loginName,
            username: loginUsername,
            password: loginPassword,
            url: loginUrl || undefined,
            notes: loginNotes || undefined,
            favorite: false,
            tags: []
          } as LoginItem;
          break;

        case 'note':
          if (!noteName || !noteContent) {
            throw new Error('Name and content are required');
          }
          itemData = {
            type: 'note',
            name: noteName,
            content: noteContent,
            favorite: false,
            tags: []
          } as NoteItem;
          break;

        case 'card':
          if (!cardName || !cardholderName || !cardNumber || !expiryMonth || !expiryYear || !cvv) {
            throw new Error('All card fields are required');
          }
          itemData = {
            type: 'card',
            name: cardName,
            cardholderName,
            cardNumber,
            expiryMonth,
            expiryYear,
            cvv,
            notes: cardNotes || undefined,
            favorite: false,
            tags: []
          } as CardItem;
          break;

        case 'identity':
          if (!identityName || !firstName || !lastName) {
            throw new Error('Name, first name, and last name are required');
          }
          itemData = {
            type: 'identity',
            name: identityName,
            firstName,
            lastName,
            email: email || undefined,
            phone: phone || undefined,
            address: address || undefined,
            notes: identityNotes || undefined,
            favorite: false,
            tags: []
          } as IdentityItem;
          break;
      }

      await onSave(itemData);
      onClose();
    } catch (err: any) {
      setError(err.message || 'Failed to save item');
    } finally {
      setLoading(false);
    }
  };

  const passwordStrength = loginPassword ? calculatePasswordStrength(loginPassword) : null;

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <div className="modal-header">
          <h2>Add New Item</h2>
          <button className="modal-close" onClick={onClose}>×</button>
        </div>

        <div className="modal-body">
          {error && <div className="error-message">{error}</div>}

          <div className="item-type-selector">
            <button
              type="button"
              className={`type-btn ${itemType === 'login' ? 'active' : ''}`}
              onClick={() => setItemType('login')}
            >
              🔑 Login
            </button>
            <button
              type="button"
              className={`type-btn ${itemType === 'note' ? 'active' : ''}`}
              onClick={() => setItemType('note')}
            >
              📝 Note
            </button>
            <button
              type="button"
              className={`type-btn ${itemType === 'card' ? 'active' : ''}`}
              onClick={() => setItemType('card')}
            >
              💳 Card
            </button>
            <button
              type="button"
              className={`type-btn ${itemType === 'identity' ? 'active' : ''}`}
              onClick={() => setItemType('identity')}
            >
              👤 Identity
            </button>
          </div>

          <form onSubmit={handleSubmit}>
            {itemType === 'login' && (
              <>
                <div className="form-group">
                  <label>Name *</label>
                  <input
                    type="text"
                    value={loginName}
                    onChange={(e) => setLoginName(e.target.value)}
                    placeholder="e.g., GitHub, Gmail"
                    required
                  />
                </div>

                <div className="form-group">
                  <label>Username *</label>
                  <input
                    type="text"
                    value={loginUsername}
                    onChange={(e) => setLoginUsername(e.target.value)}
                    placeholder="username or email"
                    required
                  />
                </div>

                <div className="form-group">
                  <label>Password *</label>
                  <div className="password-field">
                    <input
                      type="text"
                      value={loginPassword}
                      onChange={(e) => setLoginPassword(e.target.value)}
                      placeholder="password"
                      required
                    />
                    <button
                      type="button"
                      className="btn-generate"
                      onClick={() => setShowPasswordGenerator(!showPasswordGenerator)}
                    >
                      🎲
                    </button>
                  </div>
                  {passwordStrength && (
                    <div className="password-strength">
                      <div className="strength-bar">
                        <div
                          className="strength-fill"
                          style={{
                            width: `${(passwordStrength.score + 1) * 20}%`,
                            backgroundColor: passwordStrength.color
                          }}
                        />
                      </div>
                      <span style={{ color: passwordStrength.color }}>
                        {passwordStrength.label}
                      </span>
                    </div>
                  )}
                  {showPasswordGenerator && (
                    <div className="password-generator">
                      <div className="generator-controls">
                        <label>
                          Length: {passwordLength}
                          <input
                            type="range"
                            min="8"
                            max="64"
                            value={passwordLength}
                            onChange={(e) => setPasswordLength(parseInt(e.target.value))}
                          />
                        </label>
                        <label>
                          <input
                            type="checkbox"
                            checked={includeUppercase}
                            onChange={(e) => setIncludeUppercase(e.target.checked)}
                          />
                          Uppercase (A-Z)
                        </label>
                        <label>
                          <input
                            type="checkbox"
                            checked={includeLowercase}
                            onChange={(e) => setIncludeLowercase(e.target.checked)}
                          />
                          Lowercase (a-z)
                        </label>
                        <label>
                          <input
                            type="checkbox"
                            checked={includeNumbers}
                            onChange={(e) => setIncludeNumbers(e.target.checked)}
                          />
                          Numbers (0-9)
                        </label>
                        <label>
                          <input
                            type="checkbox"
                            checked={includeSymbols}
                            onChange={(e) => setIncludeSymbols(e.target.checked)}
                          />
                          Symbols (!@#$...)
                        </label>
                      </div>
                      <button type="button" onClick={handleGeneratePassword} className="btn-secondary">
                        Generate Password
                      </button>
                    </div>
                  )}
                </div>

                <div className="form-group">
                  <label>Website URL</label>
                  <input
                    type="url"
                    value={loginUrl}
                    onChange={(e) => setLoginUrl(e.target.value)}
                    placeholder="https://example.com"
                  />
                </div>

                <div className="form-group">
                  <label>Notes</label>
                  <textarea
                    value={loginNotes}
                    onChange={(e) => setLoginNotes(e.target.value)}
                    placeholder="Additional notes..."
                    rows={3}
                  />
                </div>
              </>
            )}

            {itemType === 'note' && (
              <>
                <div className="form-group">
                  <label>Name *</label>
                  <input
                    type="text"
                    value={noteName}
                    onChange={(e) => setNoteName(e.target.value)}
                    placeholder="Note title"
                    required
                  />
                </div>

                <div className="form-group">
                  <label>Content *</label>
                  <textarea
                    value={noteContent}
                    onChange={(e) => setNoteContent(e.target.value)}
                    placeholder="Your secure note..."
                    rows={10}
                    required
                  />
                </div>
              </>
            )}

            {itemType === 'card' && (
              <>
                <div className="form-group">
                  <label>Name *</label>
                  <input
                    type="text"
                    value={cardName}
                    onChange={(e) => setCardName(e.target.value)}
                    placeholder="e.g., Visa *1234"
                    required
                  />
                </div>

                <div className="form-group">
                  <label>Cardholder Name *</label>
                  <input
                    type="text"
                    value={cardholderName}
                    onChange={(e) => setCardholderName(e.target.value)}
                    placeholder="Name on card"
                    required
                  />
                </div>

                <div className="form-group">
                  <label>Card Number *</label>
                  <input
                    type="text"
                    value={cardNumber}
                    onChange={(e) => setCardNumber(e.target.value)}
                    placeholder="1234 5678 9012 3456"
                    maxLength={19}
                    required
                  />
                </div>

                <div className="form-row">
                  <div className="form-group">
                    <label>Expiry Month *</label>
                    <input
                      type="text"
                      value={expiryMonth}
                      onChange={(e) => setExpiryMonth(e.target.value)}
                      placeholder="MM"
                      maxLength={2}
                      required
                    />
                  </div>

                  <div className="form-group">
                    <label>Expiry Year *</label>
                    <input
                      type="text"
                      value={expiryYear}
                      onChange={(e) => setExpiryYear(e.target.value)}
                      placeholder="YYYY"
                      maxLength={4}
                      required
                    />
                  </div>

                  <div className="form-group">
                    <label>CVV *</label>
                    <input
                      type="text"
                      value={cvv}
                      onChange={(e) => setCvv(e.target.value)}
                      placeholder="123"
                      maxLength={4}
                      required
                    />
                  </div>
                </div>

                <div className="form-group">
                  <label>Notes</label>
                  <textarea
                    value={cardNotes}
                    onChange={(e) => setCardNotes(e.target.value)}
                    placeholder="Additional notes..."
                    rows={3}
                  />
                </div>
              </>
            )}

            {itemType === 'identity' && (
              <>
                <div className="form-group">
                  <label>Name *</label>
                  <input
                    type="text"
                    value={identityName}
                    onChange={(e) => setIdentityName(e.target.value)}
                    placeholder="e.g., Personal Identity"
                    required
                  />
                </div>

                <div className="form-row">
                  <div className="form-group">
                    <label>First Name *</label>
                    <input
                      type="text"
                      value={firstName}
                      onChange={(e) => setFirstName(e.target.value)}
                      placeholder="First name"
                      required
                    />
                  </div>

                  <div className="form-group">
                    <label>Last Name *</label>
                    <input
                      type="text"
                      value={lastName}
                      onChange={(e) => setLastName(e.target.value)}
                      placeholder="Last name"
                      required
                    />
                  </div>
                </div>

                <div className="form-group">
                  <label>Email</label>
                  <input
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    placeholder="email@example.com"
                  />
                </div>

                <div className="form-group">
                  <label>Phone</label>
                  <input
                    type="tel"
                    value={phone}
                    onChange={(e) => setPhone(e.target.value)}
                    placeholder="+1 234 567 8900"
                  />
                </div>

                <div className="form-group">
                  <label>Address</label>
                  <textarea
                    value={address}
                    onChange={(e) => setAddress(e.target.value)}
                    placeholder="Full address..."
                    rows={3}
                  />
                </div>

                <div className="form-group">
                  <label>Notes</label>
                  <textarea
                    value={identityNotes}
                    onChange={(e) => setIdentityNotes(e.target.value)}
                    placeholder="Additional notes..."
                    rows={3}
                  />
                </div>
              </>
            )}

            <div className="modal-footer">
              <button type="button" onClick={onClose} className="btn-secondary" disabled={loading}>
                Cancel
              </button>
              <button type="submit" className="btn-primary" disabled={loading}>
                {loading ? 'Saving...' : 'Save Item'}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
}
