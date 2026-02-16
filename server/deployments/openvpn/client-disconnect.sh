#!/bin/bash
# Called by OpenVPN when a client disconnects

API_URL="http://api:8080/api"

echo "Client disconnected: $common_name at $ifconfig_pool_remote_ip"

curl -s -X POST "$API_URL/internal/vpn/disconnected" \
  -H "Content-Type: application/json" \
  -d "{
    \"common_name\": \"$common_name\",
    \"vpn_ip\": \"$ifconfig_pool_remote_ip\"
  }"

exit 0
