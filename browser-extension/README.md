# RootLock Browser Extension

Browser extension for RootLock Password Manager providing secure autofill and password management.

## Features

- 🔐 Secure password autofill
- 🔄 Syncs with RootLock desktop app
- 🎯 Auto-detects login forms
- 🛡️ Zero-knowledge encryption
- ⚡ Native messaging for secure communication

## Development Setup

### Prerequisites

- Node.js 18+ 
- npm or yarn
- RootLock desktop app installed

### Installation

```bash
# Install dependencies
cd browser-extension
npm install

# Development build (with watch mode)
npm run dev

# Production build
npm run build
```

### Load Extension in Browser

#### Chrome/Edge

1. Open `chrome://extensions/`
2. Enable "Developer mode"
3. Click "Load unpacked"
4. Select the `browser-extension/dist` folder

#### Firefox

1. Open `about:debugging#/runtime/this-firefox`
2. Click "Load Temporary Add-on"
3. Select any file in the `browser-extension/dist` folder

## Project Structure

```
browser-extension/
├── src/
│   ├── popup/             # Extension popup UI
│   │   ├── Popup.tsx
│   │   ├── index.html
│   │   └── popup.css
│   ├── content/           # Content scripts (form detection)
│   │   └── autofill.ts
│   ├── background/        # Service worker
│   │   └── service-worker.ts
│   └── shared/            # Shared utilities
│       └── types.ts
├── public/                # Static assets
├── dist/                  # Build output (gitignored)
├── manifest.json          # Extension manifest
├── package.json
├── tsconfig.json
└── vite.config.ts
```

## Architecture

```
┌─────────────────────────────────────┐
│      Browser Extension              │
│  ┌──────────┐  ┌─────────────────┐ │
│  │ Popup UI │  │ Content Script  │ │
│  └────┬─────┘  └────┬────────────┘ │
│       └─────────────┼──────────────┘
│                     │
│         ┌───────────▼─────────────┐
│         │ Background Worker       │
│         └───────────┬─────────────┘
│                     │
└─────────────────────┼──────────────┘
                      │ Native Messaging
┌─────────────────────▼──────────────┐
│   RootLock Desktop App             │
└────────────────────────────────────┘
```

## Current Status

### ✅ Implemented
- Basic extension structure
- Popup UI design
- Form detection in content script
- Service worker setup
- Manifest V3 configuration

### 🚧 In Progress
- Native messaging bridge
- Desktop app communication
- Credential retrieval
- Autofill functionality

### 📋 Planned
- Password generation
- Save new credentials
- Update existing credentials
- Multi-account support
- Keyboard shortcuts
- Settings panel

## Usage

1. **Install Extension**: Load the unpacked extension in your browser
2. **Open Desktop App**: Make sure RootLock desktop app is running
3. **Navigate to Login Page**: Visit any website with a login form
4. **Auto-detect**: Extension will detect login forms automatically
5. **Click Autofill**: Click the 🔐 button next to password fields
6. **Select Account**: Choose which account to autofill
7. **Login**: Credentials are filled automatically

## Security

- Extension never stores passwords locally
- All encryption/decryption happens in desktop app
- Secure native messaging channel
- No cloud sync - everything stays on your device

## Development

### Watch Mode
```bash
npm run dev
```
Changes will automatically rebuild. Reload the extension in your browser to see updates.

### Build for Production
```bash
npm run build
```
Creates optimized build in `dist/` folder.

### Type Checking
```bash
npx tsc --noEmit
```

## Browser Support

- ✅ Chrome 88+
- ✅ Edge 88+
- ✅ Firefox 109+
- ⏳ Safari (planned)

## License

Same as RootLock desktop app - see main README.
