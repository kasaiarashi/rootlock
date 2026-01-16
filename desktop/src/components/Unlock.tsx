import { useState } from 'react';
import { useAuth } from '../contexts/AuthContext';
import './Auth.css';

export default function Unlock() {
  const [masterPassword, setMasterPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const { user, unlock, logout } = useAuth();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    try {
      await unlock(masterPassword);
    } catch (err: any) {
      setError(err.message || 'Failed to unlock vault');
    } finally {
      setLoading(false);
    }
  };

  const handleLogout = async () => {
    await logout();
  };

  return (
    <div className="auth-form">
      <div className="auth-header">
        <h1>🔐 RootLock</h1>
        <p className="auth-subtitle">Unlock Your Vault</p>
      </div>

      <div className="user-badge">
        <div className="user-avatar">
          {user?.email.charAt(0).toUpperCase()}
        </div>
        <div className="user-info-text">
          <strong>{user?.email}</strong>
          <span className="user-subtitle">Enter your master password to unlock</span>
        </div>
      </div>

      {error && <div className="error-message">{error}</div>}

      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="masterPassword">Master Password</label>
          <input
            id="masterPassword"
            type="password"
            value={masterPassword}
            onChange={(e) => setMasterPassword(e.target.value)}
            placeholder="Enter your master password"
            required
            autoFocus
            disabled={loading}
          />
        </div>

        <button type="submit" className="btn-primary" disabled={loading}>
          {loading ? 'Unlocking...' : 'Unlock Vault'}
        </button>

        <button
          type="button"
          onClick={handleLogout}
          className="btn-link"
          disabled={loading}
        >
          Logout and login with different account
        </button>
      </form>

      <div className="auth-footer">
        <p className="info-text">
          🔒 Your vault is encrypted. The secret key is stored securely on this device.
        </p>
      </div>
    </div>
  );
}
