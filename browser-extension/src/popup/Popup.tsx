import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { Lock, Key, Settings, RefreshCw } from 'lucide-react';
import './popup.css';

function Popup() {
  return (
    <div className="popup-container">
      <header className="popup-header">
        <div className="header-content">
          <Lock size={20} />
          <h1>RootLock</h1>
        </div>
      </header>

      <main className="popup-main">
        <div className="status-card">
          <div className="status-icon locked">
            <Lock size={24} />
          </div>
          <div className="status-text">
            <h2>Vault Locked</h2>
            <p>Open RootLock desktop app to unlock your vault</p>
          </div>
        </div>

        <div className="quick-actions">
          <button className="action-btn">
            <Key size={16} />
            <span>View Passwords</span>
          </button>
          <button className="action-btn">
            <RefreshCw size={16} />
            <span>Sync Vault</span>
          </button>
          <button className="action-btn">
            <Settings size={16} />
            <span>Settings</span>
          </button>
        </div>

        <div className="footer-info">
          <p>Extension v0.1.0</p>
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
