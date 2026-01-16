import { StrictMode, useState, useEffect } from 'react';
import { createRoot } from 'react-dom/client';
import { Lock, Unlock, Key, Settings, RefreshCw, AlertCircle } from 'lucide-react';
import './popup.css';

interface VaultStatus {
  connected: boolean;
  locked: boolean;
  version?: string;
}

function Popup() {
  const [status, setStatus] = useState<VaultStatus>({ connected: false, locked: true });
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Check vault status on mount
    chrome.runtime.sendMessage({ type: 'GET_STATUS' }, (response) => {
      console.log('Status response:', response);
      if (response) {
        setStatus({
          connected: response.connected ?? false,
          locked: response.locked ?? true,
          version: response.version,
        });
      }
      setLoading(false);
    });
  }, []);

  const handleRefresh = () => {
    setLoading(true);
    chrome.runtime.sendMessage({ type: 'GET_STATUS' }, (response) => {
      if (response) {
        setStatus({
          connected: response.connected ?? false,
          locked: response.locked ?? true,
          version: response.version,
        });
      }
      setLoading(false);
    });
  };

  return (
    <div className="popup-container">
      <header className="popup-header">
        <div className="header-content">
          <Lock size={20} />
          <h1>RootLock</h1>
        </div>
      </header>

      <main className="popup-main">
        {loading ? (
          <div className="status-card">
            <div className="status-icon">
              <RefreshCw size={24} className="spinning" />
            </div>
            <div className="status-text">
              <h2>Checking Status...</h2>
              <p>Connecting to desktop app</p>
            </div>
          </div>
        ) : !status.connected ? (
          <div className="status-card error">
            <div className="status-icon disconnected">
              <AlertCircle size={24} />
            </div>
            <div className="status-text">
              <h2>Not Connected</h2>
              <p>Make sure RootLock desktop app is running</p>
            </div>
          </div>
        ) : status.locked ? (
          <div className="status-card">
            <div className="status-icon locked">
              <Lock size={24} />
            </div>
            <div className="status-text">
              <h2>Vault Locked</h2>
              <p>Open RootLock desktop app to unlock your vault</p>
            </div>
          </div>
        ) : (
          <div className="status-card">
            <div className="status-icon unlocked">
              <Unlock size={24} />
            </div>
            <div className="status-text">
              <h2>Vault Unlocked</h2>
              <p>Ready to autofill passwords</p>
            </div>
          </div>
        )}

        <div className="quick-actions">
          <button className="action-btn" disabled={!status.connected || status.locked}>
            <Key size={16} />
            <span>View Passwords</span>
          </button>
          <button className="action-btn" onClick={handleRefresh}>
            <RefreshCw size={16} />
            <span>Refresh Status</span>
          </button>
          <button className="action-btn">
            <Settings size={16} />
            <span>Settings</span>
          </button>
        </div>

        <div className="footer-info">
          <p>Extension v0.1.0 {status.connected && status.version && `• Host v${status.version}`}</p>
          <a href="#" className="footer-link">Need help?</a>
        </div>
      </main>
    </div>
  );
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <Popup />
  </StrictMode>
);
