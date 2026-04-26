#!/usr/bin/env bash

set -euo pipefail

BASE_URL="${1:-${BASE_URL:-}}"

if [ -z "$BASE_URL" ]; then
  echo "Usage: BASE_URL=https://your-service.onrender.com ./scripts/smoke_e2e.sh"
  echo "Or: ./scripts/smoke_e2e.sh https://your-service.onrender.com"
  exit 1
fi

BASE_URL="${BASE_URL%/}"
ORIGINAL_URL="https://www.google.com"

require_python() {
  if ! command -v python3 >/dev/null 2>&1; then
    echo "python3 is required to parse JSON responses"
    exit 1
  fi
}

json_get() {
  local key="$1"
  python3 -c "import json,sys; data=json.load(sys.stdin); print(data.get('$key',''))"
}

require_python

echo "[1/8] healthcheck"
HEALTH_BODY=$(curl -sS "$BASE_URL/health")
echo "$HEALTH_BODY" | python3 -c "import json,sys; d=json.load(sys.stdin); assert d.get('status')=='ok', d"

echo "[2/8] create short url"
CREATE_BODY=$(curl -sS -X POST "$BASE_URL/api/v1/urls" -H "Content-Type: application/json" -d "{\"originalUrl\":\"$ORIGINAL_URL\"}")
SHORT_URL=$(printf '%s' "$CREATE_BODY" | json_get shortUrl)
if [ -z "$SHORT_URL" ]; then
  echo "create failed: $CREATE_BODY"
  exit 1
fi
CODE="${SHORT_URL##*/}"
echo "created code=$CODE"

echo "[3/8] get by code"
GET_BY_CODE_BODY=$(curl -sS "$BASE_URL/api/v1/urls/$CODE")
printf '%s' "$GET_BY_CODE_BODY" | python3 -c "import json,sys; d=json.load(sys.stdin); assert d.get('shortCode')=='$CODE', d"

echo "[4/8] redirect active"
REDIRECT_HEADERS=$(curl -sS -D - -o /dev/null "$BASE_URL/$CODE")
printf '%s' "$REDIRECT_HEADERS" | grep -q "302"
printf '%s' "$REDIRECT_HEADERS" | grep -qi "^location: $ORIGINAL_URL"

echo "[5/8] deactivate"
DEACTIVATE_BODY=$(curl -sS -X PATCH "$BASE_URL/api/v1/urls/$CODE/deactivate")
printf '%s' "$DEACTIVATE_BODY" | python3 -c "import json,sys; d=json.load(sys.stdin); assert d.get('success') is True, d"

echo "[6/8] redirect inactive"
INACTIVE_BODY=$(curl -sS "$BASE_URL/$CODE")
printf '%s' "$INACTIVE_BODY" | python3 -c "import json,sys; d=json.load(sys.stdin); assert d.get('error',{}).get('code')=='URL_NOT_ACTIVE', d"

echo "[7/8] activate"
ACTIVATE_BODY=$(curl -sS -X PATCH "$BASE_URL/api/v1/urls/$CODE/active")
printf '%s' "$ACTIVATE_BODY" | python3 -c "import json,sys; d=json.load(sys.stdin); assert d.get('success') is True, d"

echo "[8/8] expiration"
EXPIRES_AT=$(date -u -d '+20 seconds' +"%Y-%m-%dT%H:%M:%SZ")
CREATE_EXP_BODY=$(curl -sS -X POST "$BASE_URL/api/v1/urls" -H "Content-Type: application/json" -d "{\"originalUrl\":\"$ORIGINAL_URL\",\"expiresAt\":\"$EXPIRES_AT\"}")
EXP_SHORT_URL=$(printf '%s' "$CREATE_EXP_BODY" | json_get shortUrl)
EXP_CODE="${EXP_SHORT_URL##*/}"
if [ -z "$EXP_CODE" ]; then
  echo "expiration create failed: $CREATE_EXP_BODY"
  exit 1
fi
echo "waiting 25 seconds for expiration code=$EXP_CODE"
sleep 25
EXPIRED_BODY=$(curl -sS "$BASE_URL/$EXP_CODE")
printf '%s' "$EXPIRED_BODY" | python3 -c "import json,sys; d=json.load(sys.stdin); assert d.get('error',{}).get('code')=='URL_EXPIRED', d"

echo "Smoke E2E passed for $BASE_URL"
