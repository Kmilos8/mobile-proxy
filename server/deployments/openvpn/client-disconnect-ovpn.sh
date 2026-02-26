#!/bin/sh
# Called by OpenVPN (client-server) when a client disconnects.
# Exit code does not affect client state (already disconnected),
# but we still fail properly for logging visibility.
#
# Environment variables from OpenVPN:
# - username: the authenticated username
# - ifconfig_pool_remote_ip: the VPN IP that was assigned (10.9.0.x)

if [ -f /etc/openvpn/api_url ]; then
  API_URL=$(cat /etc/openvpn/api_url)
else
  API_URL="${OPENVPN_API_URL:-http://127.0.0.1:8080/api}"
fi

echo "OpenVPN client disconnected: $username at $ifconfig_pool_remote_ip"

if ! wget -q -O - --timeout=5 \
  --post-data="{\"username\":\"$username\",\"vpn_ip\":\"$ifconfig_pool_remote_ip\"}" \
  --header="Content-Type: application/json" \
  "$API_URL/internal/openvpn/disconnect"; then
  echo "WARNING: Failed to notify API of disconnect for $username — iptables cleanup may be stale"
  exit 1
fi

exit 0
