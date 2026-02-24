#!/bin/sh
# Called by OpenVPN auth-user-pass-verify via-env
# Environment variables: username, password
# Exit 0 = auth success, exit 1 = auth failure

# Read API URL from file (set during deployment), env var, or default to localhost
if [ -f /etc/openvpn/api_url ]; then
  API_URL=$(cat /etc/openvpn/api_url)
else
  API_URL="${OPENVPN_API_URL:-http://127.0.0.1:8080/api}"
fi

RESULT=$(wget -q -O - --post-data="{\"username\":\"$username\",\"password\":\"$password\"}" \
  --header="Content-Type: application/json" \
  "$API_URL/internal/openvpn/auth" 2>/dev/null)

if echo "$RESULT" | grep -q '"ok":true'; then
  echo "Auth OK for user $username"
  exit 0
else
  echo "Auth FAILED for user $username"
  exit 1
fi
