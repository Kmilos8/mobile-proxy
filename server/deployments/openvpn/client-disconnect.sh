#!/bin/sh
# Called by OpenVPN when a client disconnects

API_URL="http://127.0.0.1:8080/api"

echo "Client disconnected: $common_name at $ifconfig_pool_remote_ip"

wget -q -O - --post-data="{\"common_name\":\"$common_name\",\"vpn_ip\":\"$ifconfig_pool_remote_ip\"}" \
  --header="Content-Type: application/json" \
  "$API_URL/internal/vpn/disconnected" || true

exit 0
