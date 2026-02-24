#!/bin/bash
# One-time PKI setup for the OpenVPN client-server instance
# Run on the relay server after deploying
# Usage: setup-client-ovpn.sh <server-public-ip>

set -e

SERVER_IP="${1:?Usage: setup-client-ovpn.sh <server-public-ip>}"
VOLUME_NAME="mobile-proxy_openvpn_client_data"

echo "=== Setting up OpenVPN Client Server PKI ==="
echo "Server IP: $SERVER_IP"
echo "Volume: $VOLUME_NAME"

# Generate server config and PKI
docker run -v "$VOLUME_NAME:/etc/openvpn" --rm kylemanna/openvpn ovpn_genconfig \
  -u udp://$SERVER_IP:1195 \
  -s 10.9.0.0/24 \
  -d \
  -N \
  -C AES-256-GCM \
  -a SHA256

echo ""
echo "=== Initializing PKI (you will be prompted for a CA passphrase) ==="
docker run -v "$VOLUME_NAME:/etc/openvpn" --rm -it kylemanna/openvpn ovpn_initpki

# Generate tls-auth key
docker run -v "$VOLUME_NAME:/etc/openvpn" --rm kylemanna/openvpn openvpn --genkey --secret /etc/openvpn/pki/ta.key 2>/dev/null || true

# Create log directory
docker run -v "$VOLUME_NAME:/etc/openvpn" --rm kylemanna/openvpn mkdir -p /var/log/openvpn

# Copy our custom config over the generated one
echo ""
echo "=== Copying custom server config ==="
docker run -v "$VOLUME_NAME:/etc/openvpn" \
  -v "$(pwd)/server/deployments/openvpn/client-server.conf:/tmp/client-server.conf:ro" \
  --rm kylemanna/openvpn cp /tmp/client-server.conf /etc/openvpn/openvpn.conf

echo ""
echo "=== PKI Setup Complete ==="
echo "CA cert: inside volume $VOLUME_NAME at /etc/openvpn/pki/ca.crt"
echo "TLS auth key: inside volume $VOLUME_NAME at /etc/openvpn/pki/ta.key"
echo ""
echo "Start the OpenVPN client server with: docker compose up -d openvpn-client"
