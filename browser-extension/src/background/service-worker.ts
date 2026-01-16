// Background service worker for RootLock browser extension

import {
  initializeNativeMessaging,
  getVaultStatus,
  getCredentials,
  generatePassword,
} from '../shared/nativeMessaging';

console.log('RootLock service worker initialized');

// Initialize native messaging connection
initializeNativeMessaging();

// Handle extension installation
chrome.runtime.onInstalled.addListener((details) => {
  if (details.reason === 'install') {
    console.log('RootLock extension installed');
    // Set default settings
    chrome.storage.local.set({
      autoFillEnabled: true,
      desktopConnected: false,
    });
  } else if (details.reason === 'update') {
    console.log('RootLock extension updated');
  }
});

// Handle messages from popup and content scripts
chrome.runtime.onMessage.addListener((message, _sender, sendResponse) => {
  console.log('Received message:', message);
  
  switch (message.type) {
    case 'GET_STATUS':
      // Check desktop app connection status
      getVaultStatus()
        .then((status) => {
          sendResponse({ success: true, ...status });
        })
        .catch((error) => {
          console.error('Failed to get status:', error);
          sendResponse({ success: false, connected: false, locked: true });
        });
      break;
      
    case 'GET_CREDENTIALS':
      // Request credentials from desktop app via native messaging
      const { url, domain } = message;
      getCredentials(url, domain)
        .then((credentials) => {
          sendResponse({ success: true, credentials });
        })
        .catch((error) => {
          console.error('Failed to get credentials:', error);
          sendResponse({ success: false, credentials: [] });
        });
      break;
      
    case 'GENERATE_PASSWORD':
      // Generate password via desktop app
      generatePassword()
        .then((result) => {
          sendResponse({ success: true, password: result.password });
        })
        .catch((error) => {
          console.error('Failed to generate password:', error);
          sendResponse({ success: false, error: error.message });
        });
      break;
      
    case 'SAVE_CREDENTIAL':
      // TODO: Send new credential to desktop app
      sendResponse({ success: false, message: 'Not implemented yet' });
      break;
      
    default:
      sendResponse({ error: 'Unknown message type' });
  }
  
  return true; // Keep message channel open for async response
});

// Context menu for password generation
chrome.contextMenus.create({
  id: 'generate-password',
  title: 'Generate Password',
  contexts: ['editable'],
});

chrome.contextMenus.onClicked.addListener((info, tab) => {
  if (info.menuItemId === 'generate-password' && tab?.id) {
    // TODO: Generate password and insert into field
    chrome.tabs.sendMessage(tab.id, {
      type: 'INSERT_PASSWORD',
      password: 'GeneratedPassword123!',
    });
  }
});

export {};
