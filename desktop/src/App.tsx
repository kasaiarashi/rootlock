import { AuthProvider, useAuth } from './contexts/AuthContext';
import Login from './components/Login';
import Register from './components/Register';
import Unlock from './components/Unlock';
import VaultManager from './components/VaultManager';
import './App.css';
import { useState } from 'react';

function AppContent() {
  const { isAuthenticated, hasStoredSession } = useAuth();
  const [showRegister, setShowRegister] = useState(false);

  // User is fully authenticated (has MEK in memory)
  if (isAuthenticated) {
    return <VaultManager />;
  }

  // User has a stored session but needs to unlock (re-enter password)
  if (hasStoredSession) {
    return (
      <div className="auth-container">
        <Unlock />
      </div>
    );
  }

  // User needs to login or register
  return (
    <div className="auth-container">
      {showRegister ? (
        <Register onSwitchToLogin={() => setShowRegister(false)} />
      ) : (
        <Login onSwitchToRegister={() => setShowRegister(true)} />
      )}
    </div>
  );
}

function App() {
  return (
    <AuthProvider>
      <AppContent />
    </AuthProvider>
  );
}

export default App;
