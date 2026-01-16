#!/bin/bash

# Test token refresh flow
# This script simulates the token refresh process

API_BASE="http://localhost:8080/api/v1"

echo "=== RootLock Token Refresh Test ==="
echo ""

# Step 1: Login to get tokens
echo "Step 1: Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@test.com",
    "auk": "test_auk_hash_placeholder"
  }')

echo "Login Response:"
echo "$LOGIN_RESPONSE" | jq '.' 2>/dev/null || echo "$LOGIN_RESPONSE"
echo ""

# Extract tokens (this would fail if login fails, which is expected)
ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.tokens.access_token' 2>/dev/null)
REFRESH_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.tokens.refresh_token' 2>/dev/null)

if [ "$ACCESS_TOKEN" == "null" ] || [ -z "$ACCESS_TOKEN" ]; then
  echo "❌ Login failed (expected - we don't have the correct AUK)"
  echo "   To properly test, you need to login through the app first"
  echo ""
  echo "Alternative: Testing refresh endpoint structure..."
  echo ""
  
  # Test the endpoint structure with a fake token
  echo "Step 2: Testing refresh endpoint format..."
  REFRESH_RESPONSE=$(curl -s -X POST "$API_BASE/auth/refresh" \
    -H "Content-Type: application/json" \
    -d '{"refresh_token": "fake_token_for_testing"}')
  
  echo "Refresh Response:"
  echo "$REFRESH_RESPONSE" | jq '.' 2>/dev/null || echo "$REFRESH_RESPONSE"
  echo ""
  
  if echo "$REFRESH_RESPONSE" | jq -e '.tokens' > /dev/null 2>&1; then
    echo "✅ Refresh endpoint returns correct format with 'tokens' object"
  else
    echo "❌ Refresh endpoint does not return 'tokens' object"
  fi
  
  exit 0
fi

echo "✅ Login successful!"
echo "Access Token: ${ACCESS_TOKEN:0:20}..."
echo "Refresh Token: ${REFRESH_TOKEN:0:20}..."
echo ""

# Step 2: Test vault access with valid token
echo "Step 2: Accessing vault with valid token..."
VAULT_RESPONSE=$(curl -s -X GET "$API_BASE/vault" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "Vault Response:"
echo "$VAULT_RESPONSE" | jq '.' 2>/dev/null || echo "$VAULT_RESPONSE"
echo ""

# Step 3: Test token refresh
echo "Step 3: Refreshing token..."
REFRESH_RESPONSE=$(curl -s -X POST "$API_BASE/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\": \"$REFRESH_TOKEN\"}")

echo "Refresh Response:"
echo "$REFRESH_RESPONSE" | jq '.' 2>/dev/null || echo "$REFRESH_RESPONSE"
echo ""

# Verify response structure
if echo "$REFRESH_RESPONSE" | jq -e '.tokens.access_token' > /dev/null 2>&1; then
  NEW_ACCESS_TOKEN=$(echo "$REFRESH_RESPONSE" | jq -r '.tokens.access_token')
  echo "✅ Token refresh successful!"
  echo "New Access Token: ${NEW_ACCESS_TOKEN:0:20}..."
  echo ""
  
  # Step 4: Test vault access with new token
  echo "Step 4: Accessing vault with refreshed token..."
  NEW_VAULT_RESPONSE=$(curl -s -X GET "$API_BASE/vault" \
    -H "Authorization: Bearer $NEW_ACCESS_TOKEN")
  
  echo "Vault Response:"
  echo "$NEW_VAULT_RESPONSE" | jq '.' 2>/dev/null || echo "$NEW_VAULT_RESPONSE"
  echo ""
  
  if echo "$NEW_VAULT_RESPONSE" | jq -e '.id' > /dev/null 2>&1; then
    echo "✅ All tests passed! Token refresh flow works correctly."
  else
    echo "⚠️  Token refreshed but vault access may have issues"
  fi
else
  echo "❌ Token refresh failed or returned incorrect format"
fi

echo ""
echo "=== Test Complete ==="
