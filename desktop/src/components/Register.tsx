import { useState } from 'react';
import { useAuth } from '../contexts/AuthContext';
import './Auth.css';

interface RegisterProps {
  onSwitchToLogin: () => void;
}

export default function Register({ onSwitchToLogin }: RegisterProps) {
  const [email, setEmail] = useState('');
  const [masterPassword, setMasterPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [secretKey, setSecretKey] = useState<string | null>(null);
  const [showSecretKey, setShowSecretKey] = useState(false);
  const { register } = useAuth();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (masterPassword !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }

    if (masterPassword.length < 12) {
      setError('Master password must be at least 12 characters');
      return;
    }

    setLoading(true);

    try {
      const result = await register(email, masterPassword);
      setSecretKey(result.secretKey);
      setShowSecretKey(true);
    } catch (err: any) {
      setError(err.message || 'Registration failed');
    } finally {
      setLoading(false);
    }
  };

  const handleCopySecretKey = () => {
    if (secretKey) {
      navigator.clipboard.writeText(secretKey);
      alert('Secret key copied to clipboard!');
    }
  };

  const handleDownloadSecretKey = () => {
    if (secretKey) {
      const blob = new Blob([secretKey], { type: 'text/plain' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'rootlock-secret-key.txt';
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    }
  };

  if (showSecretKey && secretKey) {
    return (
      <div className="auth-card">
        <h1>🔐 RootLock</h1>
        <h2>Save Your Secret Key</h2>
        
        <div className="warning-box">
          <p><strong>⚠️ IMPORTANT: Save this secret key!</strong></p>
          <p>You will need this key along with your master password to access your vault.</p>
          <p>If you lose this key, you will <strong>permanently</strong> lose access to your data.</p>
        </div>

        <div className="secret-key-display">
          <code>{secretKey}</code>
        </div>

        <div className="button-group">
          <button onClick={handleCopySecretKey} className="btn-secondary">
            Copy to Clipboard
          </button>
          <button onClick={handleDownloadSecretKey} className="btn-secondary">
            Download as File
          </button>
        </div>

        <div className="auth-footer">
          <button onClick={onSwitchToLogin} className="btn-primary">
            I've Saved My Key - Go to Login
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="auth-card">
      <h1>🔐 RootLock</h1>
      <h2>Create Account</h2>
      
      <form onSubmit={handleSubmit}>
        <div className="form-group">
          <label htmlFor="email">Email</label>
          <input
            type="email"
            id="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="your@email.com"
            required
            autoFocus
          />
        </div>

        <div className="form-group">
          <label htmlFor="password">Master Password</label>
          <input
            type="password"
            id="password"
            value={masterPassword}
            onChange={(e) => setMasterPassword(e.target.value)}
            placeholder="At least 12 characters"
            minLength={12}
            required
          />
          <small>Choose a strong, memorable password</small>
        </div>

        <div className="form-group">
          <label htmlFor="confirmPassword">Confirm Master Password</label>
          <input
            type="password"
            id="confirmPassword"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            placeholder="Re-enter your password"
            minLength={12}
            required
          />
        </div>

        {error && <div className="error-message">{error}</div>}

        <button type="submit" className="btn-primary" disabled={loading}>
          {loading ? 'Creating account...' : 'Create Account'}
        </button>
      </form>

      <div className="auth-footer">
        <p>
          Already have an account?{' '}
          <button onClick={onSwitchToLogin} className="btn-link">
            Sign in
          </button>
        </p>
      </div>
    </div>
  );
}
