#!/bin/sh
# Called by OpenVPN (client-server) when a client connects
# Environment variables from OpenVPN:
# - username: the authenticated username (username-as-common-name)
# - ifconfig_pool_remote_ip: assigned VPN IP (10.9.0.x)

API_URL="${OPENVPN_API_URL:-http://127.0.0.1:8080/api}"

echo "OpenVPN client connected: $username at $ifconfig_pool_remote_ip"

wget -q -O - --post-data="{\"username\":\"$username\",\"vpn_ip\":\"$ifconfig_pool_remote_ip\"}" \
  --header="Content-Type: application/json" \
  "$API_URL/internal/openvpn/connect" || true

exit 0
