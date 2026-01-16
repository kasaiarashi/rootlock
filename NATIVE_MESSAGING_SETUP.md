# Native Messaging Setup Guide

Complete guide to set up and test the RootLock browser extension with native messaging.

## Prerequisites

- RootLock desktop app built
- Chrome or Chromium-based browser
- macOS or Linux

## Step 1: Build Native Messaging Host

```bash
cd /Users/krishnateja/Developer/Work/rootlock
./install-native-host.sh
```

This will:
- Build the native messaging host binary (Release mode)
- Install it to the correct location for Chrome
- Create the native messaging manifest

**Expected output:**
```
=== Installing RootLock Native Messaging Host ===
Building native messaging host...
Binary location: /Users/krishnateja/Developer/Work/rootlock/desktop/src-tauri/target/release/native-messaging-host
✅ Binary built successfully
Platform: macOS
Native messaging directory: /Users/krishnateja/Library/Application Support/Google/Chrome/NativeMessagingHosts
✅ Manifest created
=== Installation Complete ===
```

## Step 2: Build Browser Extension

```bash
cd browser-extension
npm run build
```

The built extension will be in `browser-extension/dist/`

## Step 3: Load Extension in Chrome

1. Open Chrome and go to `chrome://extensions/`
2. Enable **"Developer mode"** (toggle in top-right)
3. Click **"Load unpacked"**
4. Navigate to and select: `/Users/krishnateja/Developer/Work/rootlock/browser-extension/dist`
5. The extension should now appear in your extensions list

## Step 4: Get Extension ID

1. On the extensions page (`chrome://extensions/`), find the RootLock extension
2. Look for the **ID** field (e.g., `abcdefghijklmnopqrstuvwxyz123456`)
3. Copy this ID

## Step 5: Update Native Messaging Manifest

Edit the manifest file to include your extension ID:

```bash
# Open the manifest
open "/Users/krishnateja/Library/Application Support/Google/Chrome/NativeMessagingHosts/com.rootlock.native.json"
```

Replace `EXTENSION_ID_PLACEHOLDER` with your actual extension ID:

```json
{
  "name": "com.rootlock.native",
  "description": "RootLock Native Messaging Host",
  "path": "/Users/krishnateja/Developer/Work/rootlock/desktop/src-tauri/target/release/native-messaging-host",
  "type": "stdio",
  "allowed_origins": [
    "chrome-extension://YOUR_ACTUAL_EXTENSION_ID_HERE/"
  ]
}
```

**Important:** Don't forget the trailing slash `/` after the extension ID!

## Step 6: Test the Connection

### Test 1: Check Extension Popup

1. Click the RootLock extension icon in Chrome toolbar
2. The popup should open
3. Check the status:
   - **"Not Connected"** = Native host not found or manifest incorrect
   - **"Vault Locked"** = Connected successfully! ✅
   - **"Vault Unlocked"** = Connected and vault is unlocked ✅

### Test 2: Check Background Console

1. Go to `chrome://extensions/`
2. Find RootLock extension
3. Click **"service worker"** link (opens DevTools)
4. Check console for logs:

**Success:**
```
RootLock service worker initialized
[Native Messaging] Initializing...
[Native Messaging] Connecting to host: com.rootlock.native
[Native Messaging] Connected successfully
```

**Failure:**
```
[Native Messaging] Failed to connect: Error: Specified native messaging host not found.
```

### Test 3: Test on a Login Page

1. Visit any login page (e.g., `github.com/login`)
2. Look for the 🔐 button next to the password field
3. Click the 🔐 button
4. Check console (F12) for logs:

```
RootLock content script loaded
Found login forms: 1
[Background] Received message: {type: 'GET_CREDENTIALS', ...}
[Native Messaging] Sending request: {...}
[Native Messaging] Received response: {...}
```

## Troubleshooting

### Issue 1: "Not Connected" in Popup

**Possible causes:**
- Native host binary not built or not in correct location
- Manifest file not in correct directory
- Extension ID not added to manifest
- Trailing slash missing in manifest

**Solutions:**
```bash
# Check if binary exists
ls -la "/Users/krishnateja/Developer/Work/rootlock/desktop/src-tauri/target/release/native-messaging-host"

# Check if manifest exists
cat "/Users/krishnateja/Library/Application Support/Google/Chrome/NativeMessagingHosts/com.rootlock.native.json"

# Verify extension ID matches
# Compare ID from chrome://extensions/ with manifest file
```

### Issue 2: "Specified native messaging host not found"

**Solution:**
1. Verify manifest path is correct
2. Ensure binary is executable:
   ```bash
   chmod +x /Users/krishnateja/Developer/Work/rootlock/desktop/src-tauri/target/release/native-messaging-host
   ```
3. Test binary manually:
   ```bash
   echo '{"type":"getStatus","requestId":"test"}' | /Users/krishnateja/Developer/Work/rootlock/desktop/src-tauri/target/release/native-messaging-host
   ```

### Issue 3: Binary Crashes

**Check stderr logs:**
```bash
# The native host logs to stderr
# Chrome redirects stderr to: ~/Library/Application Support/Google/Chrome/NativeMessagingHosts/
```

### Issue 4: Permission Denied

**Solution:**
```bash
# Make sure binary is executable
chmod +x /Users/krishnateja/Developer/Work/rootlock/desktop/src-tauri/target/release/native-messaging-host
```

## Testing Checklist

- [ ] Native host binary built and installed
- [ ] Manifest file created with correct path
- [ ] Extension loaded in Chrome
- [ ] Extension ID copied
- [ ] Manifest updated with extension ID
- [ ] Extension popup shows "Connected" or "Vault Locked"
- [ ] Background console shows successful connection
- [ ] Login forms detected on web pages
- [ ] 🔐 button appears on password fields
- [ ] Clicking 🔐 triggers credential request

## Current Functionality

**Working:**
- ✅ Native messaging connection
- ✅ Status checking (connected/locked)
- ✅ Mock credential response
- ✅ Mock password generation
- ✅ Form detection
- ✅ Extension UI

**Not Yet Implemented:**
- ⏳ Actual vault access (currently returns mock data)
- ⏳ Real credential autofill
- ⏳ Save new credentials
- ⏳ Real password generation

## Next Steps

Once native messaging is working:

1. **Connect to Desktop Vault:**
   - Add vault access to native host
   - Check if desktop app is running
   - Get actual credentials from vault

2. **Implement Autofill:**
   - Send real credentials to content script
   - Fill username and password fields
   - Handle multiple accounts

3. **Add Save Functionality:**
   - Detect new credentials
   - Send to desktop vault
   - Update existing credentials

## Useful Commands

```bash
# Rebuild native host
cd desktop/src-tauri
cargo build --bin native-messaging-host --release

# Rebuild extension
cd browser-extension
npm run build

# Reload extension
# Go to chrome://extensions/ and click reload icon

# View native host logs (check stderr output)
# Logs appear in browser console when running

# Test native host manually
echo '{"type":"getStatus","requestId":"123"}' | ./desktop/src-tauri/target/release/native-messaging-host

# Check manifest location
ls -la "$HOME/Library/Application Support/Google/Chrome/NativeMessagingHosts/"
```

## Success Criteria

Extension is working when:
1. Popup shows "Vault Locked" or "Vault Unlocked" (not "Not Connected")
2. Background console shows successful native messaging connection
3. Clicking 🔐 button triggers message exchange
4. No errors in console

You're ready to proceed to implementing actual vault access! 🚀
