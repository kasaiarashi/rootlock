#!/bin/bash

# Helper script to corrupt the access token for testing token refresh

AUTH_FILE="$HOME/Library/Application Support/com.krishnateja.rootlock-desktop-temp/auth.json"

if [ ! -f "$AUTH_FILE" ]; then
  echo "❌ Auth file not found at: $AUTH_FILE"
  echo "Please login to the app first!"
  exit 1
fi

echo "=== Corrupting Access Token for Testing ==="
echo ""
echo "Current auth file: $AUTH_FILE"
echo ""

# Backup original
cp "$AUTH_FILE" "${AUTH_FILE}.backup"
echo "✅ Created backup: ${AUTH_FILE}.backup"

# Read current token
CURRENT_TOKEN=$(jq -r '.tokens.access_token' "$AUTH_FILE")
echo "Current access token: ${CURRENT_TOKEN:0:30}..."
echo ""

# Corrupt the token by changing a few characters in the middle
CORRUPTED_TOKEN="${CURRENT_TOKEN:0:50}CORRUPTED${CURRENT_TOKEN:59}"

# Update the file
jq --arg token "$CORRUPTED_TOKEN" '.tokens.access_token = $token' "$AUTH_FILE" > "${AUTH_FILE}.tmp"
mv "${AUTH_FILE}.tmp" "$AUTH_FILE"

echo "✅ Access token corrupted!"
echo "Corrupted token: ${CORRUPTED_TOKEN:0:30}..."
echo ""
echo "When you try to use the app now, it should:"
echo "1. Get 'invalid or expired token' error"
echo "2. Automatically call refresh token"
echo "3. Retry the operation with new token"
echo ""
echo "Check the browser console for logs starting with [Token Refresh] and [Vault Load]"
echo ""
echo "To restore the original token:"
echo "  cp \"${AUTH_FILE}.backup\" \"$AUTH_FILE\""
