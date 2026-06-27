#!/bin/bash
# Script para testear los endpoints del api-mobile
# Requiere: curl, jq (opcional para pretty print)

set -e

BASE_URL="http://localhost:8080"
JWT_TOKEN=""
REFRESH_TOKEN=""
USER_EMAIL="productor@kajve.com"
USER_PASSWORD="password123"

echo "=== API Mobile Testing Script ==="
echo "Base URL: $BASE_URL"
echo ""

# 1. Login
echo "1. Testing POST /auth/login"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"email\": \"$USER_EMAIL\",
    \"password\": \"$USER_PASSWORD\"
  }")

echo "Response: $LOGIN_RESPONSE"

# Extraer tokens (requiere jq)
if command -v jq &> /dev/null; then
  JWT_TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.access_token // empty')
  REFRESH_TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.refresh_token // empty')
  USER_ID=$(echo $LOGIN_RESPONSE | jq -r '.usuario.id // empty')
  
  if [ -z "$JWT_TOKEN" ]; then
    echo "ERROR: Could not extract JWT token from login response"
    echo "Make sure PostgreSQL is running and user exists"
    exit 1
  fi
  
  echo "JWT Token: ${JWT_TOKEN:0:50}..."
  echo "Refresh Token: ${REFRESH_TOKEN:0:50}..."
  echo "User ID: $USER_ID"
else
  echo "WARNING: jq not installed, cannot extract tokens"
  echo "Please install jq for automatic token extraction"
  exit 1
fi

echo ""
echo "2. Testing GET /lotes (list lotes)"
curl -s -X GET "$BASE_URL/lotes" \
  -H "Authorization: Bearer $JWT_TOKEN" | jq .

echo ""
echo "3. Testing POST /lotes (create lote)"
CREATE_LOTE=$(curl -s -X POST "$BASE_URL/lotes" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d "{
    \"nombre\": \"Lote Test\",
    \"descripcion\": \"Lote de prueba\",
    \"area\": 100.5
  }")

echo "Response: $CREATE_LOTE"

if command -v jq &> /dev/null; then
  LOTE_ID=$(echo $CREATE_LOTE | jq -r '.id // empty')
  if [ ! -z "$LOTE_ID" ]; then
    echo "Created Lote ID: $LOTE_ID"
    
    echo ""
    echo "4. Testing GET /lotes/{id} (get specific lote)"
    curl -s -X GET "$BASE_URL/lotes/$LOTE_ID" \
      -H "Authorization: Bearer $JWT_TOKEN" | jq .
    
    echo ""
    echo "5. Testing GET /lotes/{id}/lecturas (get lecturas)"
    curl -s -X GET "$BASE_URL/lotes/$LOTE_ID/lecturas" \
      -H "Authorization: Bearer $JWT_TOKEN" | jq .
    
    echo ""
    echo "6. Testing POST /devices/link (link device)"
    curl -s -X POST "$BASE_URL/devices/link" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer $JWT_TOKEN" \
      -d "{
        \"esp32_id\": \"ESP32-TEST-001\",
        \"provisioning_token\": \"test_token\",
        \"lote_name\": \"Lote Test\"
      }" | jq .
  fi
fi

echo ""
echo "7. Testing POST /auth/refresh (refresh token)"
curl -s -X POST "$BASE_URL/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"$REFRESH_TOKEN\"
  }" | jq .

echo ""
echo "8. Testing GET /health (health check)"
curl -s -X GET "$BASE_URL/health" | jq .

echo ""
echo "=== Tests completed ==="
