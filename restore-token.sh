#!/bin/bash

# Helper script to restore the original access token

AUTH_FILE="$HOME/Library/Application Support/com.krishnateja.rootlock-desktop-temp/auth.json"
BACKUP_FILE="${AUTH_FILE}.backup"

if [ ! -f "$BACKUP_FILE" ]; then
  echo "❌ Backup file not found at: $BACKUP_FILE"
  echo "Nothing to restore!"
  exit 1
fi

echo "=== Restoring Original Token ==="
cp "$BACKUP_FILE" "$AUTH_FILE"
echo "✅ Token restored from backup"
echo ""
cat "$AUTH_FILE" | jq '.tokens.access_token' | cut -c1-50
