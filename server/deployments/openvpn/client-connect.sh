#!/bin/sh
# Called by OpenVPN when a client connects
# Environment variables from OpenVPN:
# - common_name: client certificate CN (device name)
# - ifconfig_pool_remote_ip: assigned VPN IP
# - trusted_ip: client's real IP (WiFi IP)

API_URL="http://127.0.0.1:8080/api"

echo "Client connected: $common_name at $ifconfig_pool_remote_ip (from $trusted_ip)"

# Notify the API about the connection
wget -q -O - --post-data="{\"common_name\":\"$common_name\",\"vpn_ip\":\"$ifconfig_pool_remote_ip\"}" \
  --header="Content-Type: application/json" \
  "$API_URL/internal/vpn/connected" || true

exit 0
