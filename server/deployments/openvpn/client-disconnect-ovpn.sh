#!/bin/sh
# Called by OpenVPN (client-server) when a client disconnects
# Environment variables from OpenVPN:
# - username: the authenticated username
# - ifconfig_pool_remote_ip: the VPN IP that was assigned (10.9.0.x)

if [ -f /etc/openvpn/api_url ]; then
  API_URL=$(cat /etc/openvpn/api_url)
else
  API_URL="${OPENVPN_API_URL:-http://127.0.0.1:8080/api}"
fi

echo "OpenVPN client disconnected: $username at $ifconfig_pool_remote_ip"

wget -q -O - --post-data="{\"username\":\"$username\",\"vpn_ip\":\"$ifconfig_pool_remote_ip\"}" \
  --header="Content-Type: application/json" \
  "$API_URL/internal/openvpn/disconnect" || true

exit 0
