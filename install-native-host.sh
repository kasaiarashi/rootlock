#!/bin/bash

# Install RootLock native messaging host for Chrome/Chromium browsers

set -e

echo "=== Installing RootLock Native Messaging Host ==="
echo ""

# Build the native messaging host
echo "Building native messaging host..."
cd desktop/src-tauri
cargo build --bin native-messaging-host --release
cd ../../

# Get the absolute path to the binary
BINARY_PATH="$(pwd)/desktop/src-tauri/target/release/native-messaging-host"
echo "Binary location: $BINARY_PATH"

# Check if binary exists
if [ ! -f "$BINARY_PATH" ]; then
    echo "❌ Binary not found. Build may have failed."
    exit 1
fi

echo "✅ Binary built successfully"
echo ""

# Determine OS and set native messaging directory
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    NATIVE_MESSAGING_DIR="$HOME/Library/Application Support/Google/Chrome/NativeMessagingHosts"
    echo "Platform: macOS"
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Linux
    NATIVE_MESSAGING_DIR="$HOME/.config/google-chrome/NativeMessagingHosts"
    echo "Platform: Linux"
else
    echo "❌ Unsupported platform: $OSTYPE"
    exit 1
fi

# Create directory if it doesn't exist
mkdir -p "$NATIVE_MESSAGING_DIR"
echo "Native messaging directory: $NATIVE_MESSAGING_DIR"
echo ""

# Create the manifest file
MANIFEST_FILE="$NATIVE_MESSAGING_DIR/com.rootlock.native.json"
echo "Creating manifest: $MANIFEST_FILE"

cat > "$MANIFEST_FILE" << EOF
{
  "name": "com.rootlock.native",
  "description": "RootLock Native Messaging Host",
  "path": "$BINARY_PATH",
  "type": "stdio",
  "allowed_origins": [
    "chrome-extension://EXTENSION_ID_PLACEHOLDER/"
  ]
}
EOF

echo "✅ Manifest created"
echo ""

echo "=== Installation Complete ===" 
echo ""
echo "Next steps:"
echo "1. Build the browser extension: cd browser-extension && npm run build"
echo "2. Load extension in Chrome: chrome://extensions/ → Load unpacked → select browser-extension/dist"
echo "3. Get extension ID from chrome://extensions/"
echo "4. Update manifest with extension ID:"
echo "   Edit: $MANIFEST_FILE"
echo "   Replace 'EXTENSION_ID_PLACEHOLDER' with your actual extension ID"
echo ""
echo "Test connection:"
echo "   Open extension popup and check if 'Connected' status appears"
echo ""
