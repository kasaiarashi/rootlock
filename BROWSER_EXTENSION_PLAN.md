# RootLock Browser Extension - Implementation Plan

## Overview
Create a browser extension for RootLock that provides password autofill and secure communication with the desktop application.

## Architecture

### Components

```
┌─────────────────────────────────────────────────┐
│             Browser Extension                    │
├─────────────────────────────────────────────────┤
│                                                  │
│  ┌──────────────┐  ┌─────────────────────────┐ │
│  │   Popup UI   │  │   Content Scripts       │ │
│  │  (React/TS)  │  │   (Form detection)      │ │
│  └──────┬───────┘  └───────┬─────────────────┘ │
│         │                  │                    │
│         └──────────┬───────┘                    │
│                    │                            │
│         ┌──────────▼─────────────┐              │
│         │  Background Service    │              │
│         │  Worker (messaging)    │              │
│         └──────────┬─────────────┘              │
│                    │                            │
└────────────────────┼────────────────────────────┘
                     │
            Native Messaging
                     │
┌────────────────────▼────────────────────────────┐
│         RootLock Desktop App                    │
│  (Native Messaging Host - Rust/Tauri)          │
└─────────────────────────────────────────────────┘
```

## Features

### Phase 1: Core Functionality
- [x] Desktop app with vault management
- [ ] Browser extension manifest V3
- [ ] Native messaging bridge
- [ ] Basic autofill detection
- [ ] Popup UI for vault access

### Phase 2: Autofill
- [ ] Login form detection
- [ ] Password field recognition
- [ ] Autofill on user action
- [ ] Context menu integration
- [ ] Keyboard shortcuts

### Phase 3: Advanced Features
- [ ] Password generator in extension
- [ ] Save new credentials
- [ ] Update existing credentials
- [ ] Multi-account support per domain
- [ ] Secure notes access

## Technical Stack

### Browser Extension
- **Framework**: React + TypeScript
- **Build**: Vite + Extension plugin
- **UI**: Same design system as desktop (Lucide icons)
- **Messaging**: Chrome Extension APIs (Manifest V3)
- **Target**: Chrome, Firefox, Safari (later)

### Native Messaging
- **Protocol**: JSON over stdin/stdout
- **Host**: Rust binary (part of desktop app)
- **Security**: Message signing, request validation

## Directory Structure

```
browser-extension/
├── manifest.json           # Extension manifest (V3)
├── package.json
├── tsconfig.json
├── vite.config.ts
├── src/
│   ├── popup/             # Extension popup UI
│   │   ├── index.html
│   │   ├── Popup.tsx
│   │   └── popup.css
│   ├── content/           # Content scripts
│   │   ├── autofill.ts
│   │   ├── formDetector.ts
│   │   └── overlay.tsx
│   ├── background/        # Service worker
│   │   ├── service-worker.ts
│   │   ├── nativeMessaging.ts
│   │   └── vaultSync.ts
│   ├── shared/            # Shared utilities
│   │   ├── types.ts
│   │   ├── crypto.ts
│   │   └── storage.ts
│   └── assets/            # Icons, images
│       ├── icon-16.png
│       ├── icon-48.png
│       └── icon-128.png
└── public/
    └── native-host/       # Native messaging host config
        └── com.rootlock.browser.json
```

## Native Messaging Protocol

### Message Format
```typescript
interface ExtensionRequest {
  type: 'getCredentials' | 'saveCredential' | 'generatePassword' | 'getVault';
  requestId: string;
  timestamp: number;
  data?: any;
}

interface DesktopResponse {
  requestId: string;
  success: boolean;
  data?: any;
  error?: string;
}
```

### Message Types

1. **getCredentials**
   ```json
   {
     "type": "getCredentials",
     "requestId": "uuid",
     "timestamp": 1234567890,
     "data": {
       "url": "https://example.com",
       "domain": "example.com"
     }
   }
   ```

2. **saveCredential**
   ```json
   {
     "type": "saveCredential",
     "requestId": "uuid",
     "timestamp": 1234567890,
     "data": {
       "url": "https://example.com",
       "username": "user@example.com",
       "password": "encrypted_password",
       "name": "Example Account"
     }
   }
   ```

3. **generatePassword**
   ```json
   {
     "type": "generatePassword",
     "requestId": "uuid",
     "timestamp": 1234567890,
     "data": {
       "length": 16,
       "includeSymbols": true,
       "includeNumbers": true
     }
   }
   ```

## Security Considerations

### Communication Security
- All messages include timestamp (prevent replay attacks)
- Request IDs for correlation
- Desktop app validates extension origin
- No sensitive data cached in extension

### Vault Security
- Extension never stores master password
- Desktop app remains the source of truth
- Credentials only decrypted in desktop app
- Content scripts receive minimal data

### User Permissions
```json
{
  "permissions": [
    "activeTab",
    "storage",
    "contextMenus",
    "nativeMessaging"
  ],
  "host_permissions": [
    "<all_urls>"
  ]
}
```

## Implementation Steps

### Step 1: Native Messaging Host
1. Create Rust binary for native messaging
2. Implement JSON stdin/stdout communication
3. Register host with browser
4. Test basic message passing

### Step 2: Extension Skeleton
1. Set up Vite + React + TypeScript
2. Create manifest.json (V3)
3. Implement service worker
4. Create basic popup UI

### Step 3: Form Detection
1. Detect login forms in content script
2. Find username/password fields
3. Add autofill overlay/button
4. Handle form submission

### Step 4: Vault Integration
1. Connect to native messaging
2. Request credentials for current domain
3. Display in popup
4. Autofill on user action

### Step 5: Save New Credentials
1. Detect form submission
2. Prompt user to save
3. Send to desktop app
4. Update vault

## Development Commands

```bash
# Desktop app - build native messaging host
cd desktop/src-tauri
cargo build --bin native-messaging-host

# Browser extension - development
cd browser-extension
npm run dev        # Build with watch mode
npm run build      # Production build

# Install extension
# Chrome: Load unpacked from dist/
# Firefox: Load temporary add-on from dist/
```

## Testing Strategy

### Unit Tests
- Form detection logic
- Message formatting
- Storage utilities

### Integration Tests
- Native messaging communication
- Full autofill flow
- Credential saving

### Manual Testing
- Test on popular websites (GitHub, Google, etc.)
- Test form variations
- Test error scenarios
- Test across browsers

## Browser Compatibility

### Chrome/Edge (Manifest V3)
- Primary target
- Full feature support

### Firefox (Manifest V3)
- Secondary target
- Most features compatible

### Safari
- Future consideration
- Requires Safari extension conversion

## Next Steps

1. ✅ Desktop app with vault (DONE)
2. ✅ Token refresh (DONE)
3. ✅ Touch ID (DONE - needs polish)
4. ⏳ Browser extension setup (NEXT)
5. ⏳ Native messaging bridge
6. ⏳ Basic autofill
7. ⏳ Full feature set

## Timeline Estimate

- Week 1: Native messaging + Extension skeleton (2-3 days)
- Week 2: Form detection + Basic autofill (2-3 days)
- Week 3: Popup UI + Vault integration (2-3 days)
- Week 4: Save credentials + Polish (2-3 days)
- Week 5: Testing + Bug fixes (2-3 days)

**Total: 4-5 weeks for MVP browser extension**

## Resources

- [Chrome Extension Docs](https://developer.chrome.com/docs/extensions/)
- [Native Messaging](https://developer.chrome.com/docs/extensions/mv3/nativeMessaging/)
- [Manifest V3 Migration](https://developer.chrome.com/docs/extensions/mv3/intro/)
- [Firefox Extension Docs](https://developer.mozilla.org/en-US/docs/Mozilla/Add-ons/WebExtensions)
