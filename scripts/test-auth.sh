#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "Register..."
curl -s -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123","display_name":"Test"}'
echo ""

echo "Login..."
LOGIN=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}')
echo "$LOGIN"

echo ""
echo "Me..."
ACCESS=$(python3 -c "import sys,json; print(json.loads('$LOGIN')['access_token'])")
curl -s -X GET "$BASE_URL/api/v1/auth/me" \
  -H "Authorization: Bearer $ACCESS"
echo ""

echo "Refresh..."
REFRESH=$(python3 -c "import sys,json; print(json.loads('$LOGIN')['refresh_token'])")
curl -s -X POST "$BASE_URL/api/v1/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH\"}"
echo ""
