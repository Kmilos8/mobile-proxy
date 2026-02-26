#!/bin/sh
# Called by OpenVPN (client-server) when a client connects.
# OpenVPN 2.4 blocks on this script: exit 0 = admit, exit 1 = reject.
#
# Environment variables from OpenVPN:
# - username: the authenticated username (username-as-common-name)
# - ifconfig_pool_remote_ip: assigned VPN IP (10.9.0.x)

if [ -f /etc/openvpn/api_url ]; then
  API_URL=$(cat /etc/openvpn/api_url)
else
  API_URL="${OPENVPN_API_URL:-http://127.0.0.1:8080/api}"
fi

echo "OpenVPN client connected: $username at $ifconfig_pool_remote_ip"

# Notify API to set up transparent proxy mapping + iptables REDIRECT.
# Retry once on failure (covers transient network errors).
# If both attempts fail, exit 1 so OpenVPN rejects the client.
# This is correct: admitting a client without the REDIRECT rule means
# the client connects but has no internet (traffic dropped or unproxied).
MAX_RETRIES=2
attempt=1
while [ "$attempt" -le "$MAX_RETRIES" ]; do
  if wget -q -O - --timeout=5 \
    --post-data="{\"username\":\"$username\",\"vpn_ip\":\"$ifconfig_pool_remote_ip\"}" \
    --header="Content-Type: application/json" \
    "$API_URL/internal/openvpn/connect"; then
    echo "API notified successfully for $username (attempt $attempt)"
    exit 0
  fi
  echo "WARNING: API call failed for $username (attempt $attempt/$MAX_RETRIES)"
  attempt=$((attempt + 1))
  [ "$attempt" -le "$MAX_RETRIES" ] && sleep 1
done

echo "ERROR: Failed to notify API for $username after $MAX_RETRIES attempts — rejecting client"
exit 1
