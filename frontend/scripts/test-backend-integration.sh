#!/usr/bin/env bash
set -euo pipefail

API_BASE="${API_BASE:-http://localhost:8080/api}"
COOKIE_JAR="$(mktemp)"
trap 'rm -f "$COOKIE_JAR"' EXIT

echo "[integration] checking backend health at $API_BASE/health"
curl -fsS "$API_BASE/health" >/dev/null

echo "[integration] fetching csrf token"
CSRF_JSON="$(curl -fsS -c "$COOKIE_JAR" "$API_BASE/csrf")"
CSRF_TOKEN="$(node -e 'const data = JSON.parse(process.argv[1]); process.stdout.write(data.csrfToken || "");' "$CSRF_JSON")"

if [[ -z "$CSRF_TOKEN" ]]; then
  echo "[integration] failed: missing csrf token"
  exit 1
fi

USERNAME="it_user_$(date +%s)"
PASSWORD="integration123"
PAYLOAD="{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}"

echo "[integration] registering user $USERNAME"
HTTP_CODE="$(curl -sS -o /tmp/rf_register_resp.json -w "%{http_code}" \
  -b "$COOKIE_JAR" -c "$COOKIE_JAR" \
  -H "Content-Type: application/json" \
  -H "X-CSRF-Token: $CSRF_TOKEN" \
  -X POST "$API_BASE/register" \
  -d "$PAYLOAD")"

if [[ "$HTTP_CODE" != "201" ]]; then
  echo "[integration] failed: expected 201 from /register, got $HTTP_CODE"
  echo "[integration] response: $(cat /tmp/rf_register_resp.json)"
  exit 1
fi

echo "[integration] success: frontend-backend register flow works"
